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
	if _, err := NuevoPerfil("e2", "", map[string]string{}, time.Now()); err == nil {
		t.Fatal("aceptó un nombre vacío")
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
	if ultima.ReglaID != "RA05" || ultima.NivelSiguiente != 2 || Siguiente(p, h, &ultima).Nivel != 2 {
		t.Fatalf("no aumentó dificultad: %+v", ultima)
	}
}
