package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"tesis/adaptativo/backend/domain"
)

type sesion struct {
	Perfil *domain.PerfilEstudiante
	Historial []domain.Interaccion
	Actual domain.Actividad
	Pendiente bool
	Decision *domain.DecisionAdaptativa
	Feedback string
}

type Server struct { mu sync.Mutex; sesiones map[string]*sesion }

func New() *Server { return &Server{sesiones: map[string]*sesion{}} }

type actividadPublica struct {
	ID string `json:"id"`
	Tema string `json:"tema"`
	Nivel int `json:"nivel"`
	Enunciado string `json:"enunciado"`
	Opciones []string `json:"opciones"`
}

func publica(a domain.Actividad) actividadPublica { return actividadPublica{a.ID, a.Tema, a.Nivel, a.Enunciado, a.Opciones} }

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) { enviar(w, http.StatusOK, map[string]string{"estado":"ok"}) })
	mux.HandleFunc("GET /api/diagnostic", func(w http.ResponseWriter, _ *http.Request) { enviar(w, http.StatusOK, domain.Diagnostico) })
	mux.HandleFunc("POST /api/sessions", s.crear)
	mux.HandleFunc("GET /api/sessions/{id}", s.ver)
	mux.HandleFunc("POST /api/sessions/{id}/answer", s.responder)
	mux.HandleFunc("POST /api/sessions/{id}/next", s.siguiente)
	return mux
}

func enviar(w http.ResponseWriter, estado int, valor any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(valor)
}

func errorJSON(w http.ResponseWriter, estado int, err error) { enviar(w, estado, map[string]string{"error":err.Error()}) }

func leer(w http.ResponseWriter, r *http.Request, destino any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	dec := json.NewDecoder(r.Body); dec.DisallowUnknownFields()
	if err := dec.Decode(destino); err != nil { return errors.New("JSON inválido: " + err.Error()) }
	var extra any
	if err := dec.Decode(&extra); err != io.EOF { return errors.New("envíe un solo objeto JSON") }
	return nil
}

func nuevoID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil { return "", err }
	return hex.EncodeToString(b), nil
}

func (s *Server) crear(w http.ResponseWriter, r *http.Request) {
	var entrada struct { Nombre string `json:"nombre"`; Respuestas map[string]string `json:"respuestas"` }
	if err := leer(w, r, &entrada); err != nil { errorJSON(w, 400, err); return }
	id, err := nuevoID(); if err != nil { errorJSON(w, 500, err); return }
	perfil, err := domain.NuevoPerfil(id, entrada.Nombre, entrada.Respuestas, time.Now().UTC())
	if err != nil { errorJSON(w, 400, err); return }
	sesion := &sesion{Perfil: perfil, Pendiente: true}
	sesion.Actual = domain.Siguiente(perfil, nil, nil)
	s.mu.Lock(); s.sesiones[id] = sesion; s.mu.Unlock()
	enviar(w, 201, vista(sesion))
}

func (s *Server) buscar(w http.ResponseWriter, r *http.Request) *sesion {
	id := strings.TrimSpace(r.PathValue("id"))
	x := s.sesiones[id]
	if x == nil { errorJSON(w, 404, errors.New("sesión no encontrada")) }
	return x
}

func (s *Server) ver(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock(); defer s.mu.Unlock()
	if x := s.buscar(w, r); x != nil { enviar(w, 200, vista(x)) }
}

func (s *Server) responder(w http.ResponseWriter, r *http.Request) {
	var entrada struct { ActividadID string `json:"actividadId"`; Respuesta string `json:"respuesta"`; DuracionMS int64 `json:"duracionMs"` }
	if err := leer(w, r, &entrada); err != nil { errorJSON(w, 400, err); return }
	s.mu.Lock(); defer s.mu.Unlock()
	x := s.buscar(w, r); if x == nil { return }
	if !x.Pendiente || entrada.ActividadID != x.Actual.ID { errorJSON(w, 409, errors.New("actividad ya respondida o desactualizada")); return }
	interaccion, decision, feedback, err := domain.Procesar(x.Perfil, x.Historial, x.Actual, entrada.Respuesta, entrada.DuracionMS, time.Now().UTC())
	if err != nil { errorJSON(w, 400, err); return }
	x.Historial = append(x.Historial, interaccion)
	x.Decision, x.Feedback, x.Pendiente = &decision, feedback, false
	enviar(w, 200, vista(x))
}

func (s *Server) siguiente(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock(); defer s.mu.Unlock()
	x := s.buscar(w, r); if x == nil { return }
	if x.Pendiente { errorJSON(w, 409, errors.New("responda la actividad actual primero")); return }
	x.Actual = domain.Siguiente(x.Perfil, x.Historial, x.Decision)
	x.Pendiente, x.Feedback = true, ""
	enviar(w, 200, vista(x))
}

func vista(x *sesion) any {
	var total, aciertos int
	for _, i := range x.Historial { total++; if i.Correcta { aciertos++ } }
	return struct {
		Perfil *domain.PerfilEstudiante `json:"perfil"`
		Actividad actividadPublica `json:"actividad"`
		Pendiente bool `json:"pendiente"`
		Decision *domain.DecisionAdaptativa `json:"decision,omitempty"`
		Feedback string `json:"feedback,omitempty"`
		Ultima *domain.Interaccion `json:"ultima,omitempty"`
		Progreso map[string]int `json:"progreso"`
	}{x.Perfil, publica(x.Actual), x.Pendiente, x.Decision, x.Feedback, ultima(x.Historial), map[string]int{"respondidas":total,"aciertos":aciertos}}
}

func ultima(h []domain.Interaccion) *domain.Interaccion { if len(h)==0 { return nil }; x:=h[len(h)-1]; return &x }
