package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFlujoHTTPYRespuestaDuplicada(t *testing.T) {
	h := New().Handler()
	peticion := func(metodo, ruta, cuerpo string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(metodo, ruta, bytes.NewBufferString(cuerpo))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	d := peticion(http.MethodGet, "/api/diagnostic", "")
	if d.Code != 200 || bytes.Contains(d.Body.Bytes(), []byte("\"respuesta\"")) {
		t.Fatalf("diagnóstico expuso respuestas: %s", d.Body.String())
	}
	creada := peticion(http.MethodPost, "/api/sessions", `{"nombre":"Daniel","respuestas":{"d1":"manzana","d2":"casa","d3":"am","d4":"goes"}}`)
	if creada.Code != 201 {
		t.Fatalf("sesión: %s", creada.Body.String())
	}
	var vista struct {
		Perfil struct {
			ID          string `json:"id"`
			Version     int    `json:"version"`
			Diagnostico map[string]struct {
				Aciertos       int     `json:"aciertos"`
				Total          int     `json:"total"`
				DominioInicial float64 `json:"dominioInicial"`
			} `json:"diagnostico"`
		} `json:"perfil"`
		Actividad struct {
			ID string `json:"id"`
		} `json:"actividad"`
	}
	if err := json.Unmarshal(creada.Body.Bytes(), &vista); err != nil {
		t.Fatal(err)
	}
	if vista.Perfil.Version != 1 || vista.Perfil.Diagnostico["vocabulario"].Aciertos != 2 || vista.Perfil.Diagnostico["gramatica"].Total != 2 {
		t.Fatalf("diagnóstico ausente en la API: %+v", vista.Perfil)
	}
	contenido, _ := json.Marshal(map[string]any{"actividadId": vista.Actividad.ID, "respuesta": "water", "duracionMs": 3000})
	ruta := "/api/sessions/" + vista.Perfil.ID
	primera := peticion(http.MethodPost, ruta+"/answer", string(contenido))
	if primera.Code != 200 {
		t.Fatalf("respuesta: %s", primera.Body.String())
	}
	var actualizada struct {
		Perfil struct {
			Version     int `json:"version"`
			Diagnostico map[string]struct {
				DominioInicial float64 `json:"dominioInicial"`
			} `json:"diagnostico"`
		} `json:"perfil"`
	}
	if err := json.Unmarshal(primera.Body.Bytes(), &actualizada); err != nil {
		t.Fatal(err)
	}
	if actualizada.Perfil.Version != 2 || actualizada.Perfil.Diagnostico["vocabulario"].DominioInicial != vista.Perfil.Diagnostico["vocabulario"].DominioInicial {
		t.Fatal("la respuesta no conservó el resultado diagnóstico inicial")
	}
	if peticion(http.MethodPost, ruta+"/answer", string(contenido)).Code != 409 {
		t.Fatal("permitió una respuesta duplicada")
	}
	if historial := peticion(http.MethodGet, ruta+"/interactions", ""); historial.Code != 200 || !bytes.Contains(historial.Body.Bytes(), []byte(vista.Actividad.ID)) {
		t.Fatalf("historial incompleto: %s", historial.Body.String())
	}
	if peticion(http.MethodPost, ruta+"/next", "").Code != 200 {
		t.Fatal("no avanzó")
	}
}
