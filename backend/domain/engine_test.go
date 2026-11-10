package domain

import (
	"testing"
	"time"
)

func perfilPrueba(t *testing.T) *PerfilEstudiante {
	t.Helper()
	p, err := NuevoPerfil("e1", "Daniel", map[string]string{"d1": "manzana", "d2": "casa", "d3": "am", "d4": "goes"}, time.Unix(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDiagnosticoInicial(t *testing.T) {
	p := perfilPrueba(t)
	if !p.Diagnosticado || p.Temas["vocabulario"].Dominio != 0.5 || p.Temas["gramatica"].Nivel != 1 {
		t.Fatalf("perfil inicial inesperado: %+v", p)
	}
	if r := p.Diagnostico["vocabulario"]; r.Aciertos != 2 || r.Total != 2 || r.DominioInicial != 0.5 || !r.EvaluadoEn.Equal(time.Unix(0, 0)) {
		t.Fatalf("resultado diagnóstico inesperado: %+v", r)
	}
	if _, err := NuevoPerfil("e2", "", map[string]string{}, time.Now()); err == nil {
		t.Fatal("aceptó un nombre vacío")
	}
}

func TestDiagnosticoPorTemaYValidacion(t *testing.T) {
	respuestas := map[string]string{"d1": "manzana", "d2": "perro", "d3": "is", "d4": "going"}
	p, err := NuevoPerfil("e2", " Ana ", respuestas, time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if p.Nombre != "Ana" || p.Diagnostico["vocabulario"].Aciertos != 1 || p.Diagnostico["gramatica"].Aciertos != 0 || p.Temas["vocabulario"].Dominio != 0.375 || p.Temas["gramatica"].Dominio != 0.25 {
		t.Fatalf("perfil por tema inesperado: %+v", p)
	}
	delete(respuestas, "d4")
	if _, err := NuevoPerfil("e3", "Ana", respuestas, time.Now()); err == nil {
		t.Fatal("aceptó diagnóstico incompleto")
	}
	respuestas["d4"] = "goes"
	respuestas["d5"] = "extra"
	if _, err := NuevoPerfil("e3", "Ana", respuestas, time.Now()); err == nil {
		t.Fatal("aceptó pregunta desconocida")
	}
}

func TestErrorRecurrenteCambiaApoyoYProgramaRefuerzo(t *testing.T) {
	p := perfilPrueba(t)
	var h []Interaccion
	var ultima DecisionAdaptativa
	for n, id := range []string{"v1a", "v1b"} {
		a, _ := BuscarActividad(id)
		i, d, _, err := Procesar(p, h, a, "chairIfInvalid", 5000, time.Unix(int64(n+1), 0))
		if err == nil {
			t.Fatal("aceptó opción inexistente")
		}
		respuesta := "chair"
		if id == "v1b" {
			respuesta = "shoe"
		}
		i, d, _, err = Procesar(p, h, a, respuesta, 5000, time.Unix(int64(n+1), 0))
		if err != nil {
			t.Fatal(err)
		}
		h = append(h, i)
		ultima = d
	}
	if ultima.ReglaID != "RA04" || h[1].Ayuda == h[0].Ayuda || !p.Temas["vocabulario"].Refuerzo || p.Version != 3 {
		t.Fatalf("adaptación inesperada: %+v %+v", ultima, p.Temas["vocabulario"])
	}
	if Siguiente(p, h, &ultima).Tema != "vocabulario" {
		t.Fatal("no conservó el tema para refuerzo")
	}
}

func TestAciertosRapidosSubenUnNivel(t *testing.T) {
	p := perfilPrueba(t)
	var h []Interaccion
	var ultima DecisionAdaptativa
	for n, id := range []string{"v1a", "v1b"} {
		a, _ := BuscarActividad(id)
		i, d, _, err := Procesar(p, h, a, a.Respuesta, 4000, time.Unix(int64(n+1), 0))
		if err != nil {
			t.Fatal(err)
		}
		h = append(h, i)
		ultima = d
	}
	if ultima.ReglaID != "RA06" || ultima.NivelSiguiente != 2 || Siguiente(p, h, &ultima).Nivel != 2 {
		t.Fatalf("no aumentó dificultad: %+v", ultima)
	}
}

func TestDificultadElevadaReduceUnNivel(t *testing.T) {
	p := perfilPrueba(t)
	p.Temas["vocabulario"].Nivel = 2
	var h []Interaccion
	var ultima DecisionAdaptativa
	for n, caso := range []struct{ id, respuesta string }{{"v2a", "late"}, {"v2b", "station"}} {
		a, _ := BuscarActividad(caso.id)
		i, d, _, err := Procesar(p, h, a, caso.respuesta, 4000, time.Unix(int64(n+1), 0))
		if err != nil {
			t.Fatal(err)
		}
		h = append(h, i)
		ultima = d
	}
	if ultima.ReglaID != "RA05" || ultima.NivelSiguiente != 1 || h[1].Ayuda == h[0].Ayuda || Siguiente(p, h, &ultima).Nivel != 1 {
		t.Fatalf("la dificultad elevada no redujo el nivel con apoyo alternativo: %+v", ultima)
	}
}

func TestPrecisionAltaSinDominioAltoMantieneNivel(t *testing.T) {
	p, err := NuevoPerfil("e1", "Daniel", map[string]string{"d1": "mesa", "d2": "perro", "d3": "is", "d4": "go"}, time.Unix(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	var h []Interaccion
	var ultima DecisionAdaptativa
	for n, id := range []string{"v1a", "v1b"} {
		a, _ := BuscarActividad(id)
		i, d, _, err := Procesar(p, h, a, a.Respuesta, 4000, time.Unix(int64(n+1), 0))
		if err != nil {
			t.Fatal(err)
		}
		h = append(h, i)
		ultima = d
	}
	if ultima.ReglaID == "RA06" || ultima.NivelSiguiente != 1 || p.Temas["vocabulario"].Dominio >= Umbrales.DominioAlto {
		t.Fatalf("aumentó dificultad sin dominio alto: %+v", ultima)
	}
}

func TestDominioAltoAvanzaDeTemaSinCambiarDiagnostico(t *testing.T) {
	p := perfilPrueba(t)
	var h []Interaccion
	var ultima DecisionAdaptativa
	for n, id := range []string{"v1a", "v1b", "v2a"} {
		a, _ := BuscarActividad(id)
		i, d, _, err := Procesar(p, h, a, a.Respuesta, 4000, time.Unix(int64(n+1), 0))
		if err != nil {
			t.Fatal(err)
		}
		h = append(h, i)
		ultima = d
	}
	siguiente := Siguiente(p, h, &ultima)
	if ultima.ReglaID != "RA06" || siguiente.Tema != "gramatica" || p.Diagnostico["vocabulario"].DominioInicial != 0.5 || p.Temas["vocabulario"].Dominio <= 0.75 {
		t.Fatalf("avance o trazabilidad inesperados: %+v %+v", ultima, p)
	}
}

func TestDominioBajoUsaReglaDeRefuerzo(t *testing.T) {
	p := perfilPrueba(t)
	var h []Interaccion
	var ultima DecisionAdaptativa
	for n, caso := range []struct{ id, respuesta string }{{"v1a", "chair"}, {"v1b", "shoe"}, {"v1a", "water"}} {
		a, _ := BuscarActividad(caso.id)
		i, d, _, err := Procesar(p, h, a, caso.respuesta, 4000, time.Unix(int64(n+1), 0))
		if err != nil {
			t.Fatal(err)
		}
		h = append(h, i)
		ultima = d
	}
	if ultima.ReglaID != "RA07" || !p.Temas["vocabulario"].Refuerzo {
		t.Fatalf("regla de refuerzo inesperada: %+v", ultima)
	}
}
