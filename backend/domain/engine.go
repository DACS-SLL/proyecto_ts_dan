package domain

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type DiagnosticoPregunta struct {
	ID string `json:"id"`
	Tema string `json:"tema"`
	Enunciado string `json:"enunciado"`
	Opciones []string `json:"opciones"`
	Respuesta string `json:"-"`
}

var Diagnostico = []DiagnosticoPregunta{
	{"d1", "vocabulario", "Choose the meaning of 'apple'.", []string{"manzana", "mesa", "correr"}, "manzana"},
	{"d2", "vocabulario", "Choose the meaning of 'house'.", []string{"casa", "perro", "libro"}, "casa"},
	{"d3", "gramatica", "I ___ happy.", []string{"am", "is", "are"}, "am"},
	{"d4", "gramatica", "She ___ to school.", []string{"goes", "go", "going"}, "goes"},
}

func NuevoPerfil(id, nombre string, respuestas map[string]string, ahora time.Time) (*PerfilEstudiante, error) {
	if strings.TrimSpace(nombre) == "" { return nil, errors.New("el nombre es obligatorio") }
	aciertos := map[string]int{"vocabulario": 0, "gramatica": 0}
	for _, p := range Diagnostico {
		r, ok := respuestas[p.ID]
		if !ok { return nil, fmt.Errorf("falta la respuesta %s", p.ID) }
		if !opcionValida(p.Opciones, r) { return nil, fmt.Errorf("respuesta inválida para %s", p.ID) }
		if r == p.Respuesta { aciertos[p.Tema]++ }
	}
	perfil := &PerfilEstudiante{ID: id, Nombre: strings.TrimSpace(nombre), Version: 1, Diagnosticado: true, Temas: map[string]*EstadoTema{}, ActualizadoEn: ahora}
	for _, tema := range []string{"vocabulario", "gramatica"} {
		// La prueba corta solo fija un punto de partida: nunca declara dominio alto.
		dominio := 0.25 + 0.125*float64(aciertos[tema])
		perfil.Temas[tema] = &EstadoTema{Tema: tema, Dominio: dominio, Nivel: 1, Evidencias: 2, Refuerzo: false, ActualizadoEn: ahora}
	}
	return perfil, nil
}

func opcionValida(opciones []string, respuesta string) bool {
	for _, o := range opciones { if o == respuesta { return true } }
	return false
}

func Procesar(perfil *PerfilEstudiante, historial []Interaccion, actividad Actividad, respuesta string, duracionMS int64, ahora time.Time) (Interaccion, DecisionAdaptativa, string, error) {
	if perfil == nil || !perfil.Diagnosticado { return Interaccion{}, DecisionAdaptativa{}, "", errors.New("primero complete el diagnóstico") }
	if _, ok := perfil.Temas[actividad.Tema]; !ok { return Interaccion{}, DecisionAdaptativa{}, "", errors.New("tema desconocido") }
	if !opcionValida(actividad.Opciones, respuesta) { return Interaccion{}, DecisionAdaptativa{}, "", errors.New("seleccione una opción válida") }
	if duracionMS < 0 || duracionMS > 3600000 { return Interaccion{}, DecisionAdaptativa{}, "", errors.New("duración fuera de rango") }
	correcta := respuesta == actividad.Respuesta
	previos := filtrar(historial, actividad.Tema, Umbrales.Ventana-1)
	intento := 1
	for _, anterior := range historial { if anterior.ActividadID == actividad.ID { intento++ } }
	interaccion := Interaccion{ID: fmt.Sprintf("i-%d-%d", ahora.UnixNano(), len(historial)+1), ActividadID: actividad.ID, Tema: actividad.Tema, Dificultad: actividad.Nivel, Respuesta: respuesta, Correcta: correcta, DuracionMS: duracionMS, Intento: intento, Momento: ahora}
	ventana := append(previos, interaccion)
	aciertos, erroresConsecutivos := 0, 0
	for _, x := range ventana { if x.Correcta { aciertos++ } }
	for i := len(ventana)-1; i >= 0 && !ventana[i].Correcta; i-- { erroresConsecutivos++ }
	precision := float64(aciertos) / float64(len(ventana))
	estado := perfil.Temas[actividad.Tema]
	// Media exponencial sencilla; la precisión reciente evita depender de una única respuesta.
	estado.Dominio = math.Round((0.6*estado.Dominio+0.4*precision)*100) / 100
	estado.Evidencias++
	estado.ActualizadoEn = ahora
	estado.Refuerzo = estado.Dominio < Umbrales.Refuerzo && estado.Evidencias >= 4
	perfil.Version++
	perfil.ActualizadoEn = ahora
	decision := DecisionAdaptativa{ReglaID: "RA02", TemaSiguiente: actividad.Tema, NivelSiguiente: estado.Nivel, PrecisionReciente: precision, ErroresConsecutivos: erroresConsecutivos, Dominio: estado.Dominio, InteraccionID: interaccion.ID}
	feedback := "Correcto. " + actividad.Explicacion
	if !correcta {
		feedback = "Respuesta esperada: " + actividad.Respuesta + ". " + actividad.Explicacion
		decision.ReglaID = "RA03"
		decision.Razon = "Primer error reciente: practicar el mismo tema con corrección."
		interaccion.Ayuda = "corrección y ejercicio similar"
		if erroresConsecutivos >= Umbrales.ErrorRecurrente {
			decision.ReglaID = "RA04"
			decision.NivelSiguiente = max(1, actividad.Nivel-1)
			decision.Razon = "Error recurrente: usar otro ejemplo y reducir la dificultad si es posible."
			feedback += " Otra forma de verlo: identifica primero la función de la palabra en la oración."
			interaccion.Ayuda = "explicación alternativa y menor dificultad"
		}
	} else if len(ventana) >= 2 && precision >= Umbrales.DominioAlto && duracionMS <= Umbrales.TiempoRapidoMS {
		decision.ReglaID = "RA05"
		decision.NivelSiguiente = min(3, actividad.Nivel+1)
		decision.Razon = "Precisión reciente alta y respuesta rápida: subir un nivel permitido."
	} else {
		decision.Razon = "Mantener el nivel mientras se reúne más evidencia."
	}
	if estado.Refuerzo && correcta && decision.ReglaID != "RA05" {
		decision.ReglaID = "RA06"
		decision.Razon = "El dominio sigue bajo el umbral: mantener el tema para refuerzo."
	}
	estado.Nivel = decision.NivelSiguiente
	return interaccion, decision, feedback, nil
}

func filtrar(historial []Interaccion, tema string, limite int) []Interaccion {
	var resultado []Interaccion
	for i := len(historial)-1; i >= 0 && len(resultado) < limite; i-- {
		if historial[i].Tema == tema { resultado = append(resultado, historial[i]) }
	}
	for i, j := 0, len(resultado)-1; i < j; i, j = i+1, j-1 { resultado[i], resultado[j] = resultado[j], resultado[i] }
	return resultado
}

func Siguiente(perfil *PerfilEstudiante, historial []Interaccion, decision *DecisionAdaptativa) Actividad {
	tema, nivel := "vocabulario", 1
	if decision != nil { tema, nivel = decision.TemaSiguiente, decision.NivelSiguiente } else if perfil != nil && perfil.Temas["gramatica"].Dominio < perfil.Temas["vocabulario"].Dominio { tema = "gramatica" }
	if decision != nil && decision.ReglaID == "RA05" && perfil.Temas[tema].Dominio >= Umbrales.DominioAlto && perfil.Temas[tema].Evidencias >= 5 {
		otro := "gramatica"; if tema == otro { otro = "vocabulario" }
		if perfil.Temas[otro].Dominio < perfil.Temas[tema].Dominio { tema, nivel = otro, perfil.Temas[otro].Nivel; decision.ReglaID, decision.TemaSiguiente, decision.NivelSiguiente, decision.Razon = "RA07", tema, nivel, "Dominio alto sostenido: avanzar al otro tema." }
	}
	uso := map[string]int{}
	for _, x := range historial { uso[x.ActividadID]++ }
	var elegida Actividad
	mejor := int(^uint(0) >> 1)
	for _, a := range Actividades {
		if a.Tema == tema && a.Nivel == nivel && uso[a.ID] < mejor { elegida, mejor = a, uso[a.ID] }
	}
	return elegida
}
