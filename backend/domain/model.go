package domain

import "time"

type ConfigUmbrales struct {
	ErrorRecurrente int     `json:"errorRecurrente"`
	DominioAlto     float64 `json:"dominioAlto"`
	Refuerzo        float64 `json:"refuerzo"`
	TiempoRapidoMS  int64   `json:"tiempoRapidoMs"`
	Ventana         int     `json:"ventana"`
}

var Umbrales = ConfigUmbrales{ErrorRecurrente: 2, DominioAlto: 0.75, Refuerzo: 0.5, TiempoRapidoMS: 12000, Ventana: 4}

type EstadoTema struct {
	Tema           string    `json:"tema"`
	Dominio        float64   `json:"dominio"`
	Nivel          int       `json:"nivel"`
	Evidencias     int       `json:"evidencias"`
	Refuerzo       bool      `json:"refuerzo"`
	ActualizadoEn  time.Time `json:"actualizadoEn"`
}

type PerfilEstudiante struct {
	ID            string                 `json:"id"`
	Nombre        string                 `json:"nombre"`
	Version       int                    `json:"version"`
	Diagnosticado  bool                   `json:"diagnosticado"`
	Temas         map[string]*EstadoTema `json:"temas"`
	ActualizadoEn time.Time              `json:"actualizadoEn"`
}

type Interaccion struct {
	ID            string    `json:"id"`
	ActividadID   string    `json:"actividadId"`
	Tema          string    `json:"tema"`
	Dificultad    int       `json:"dificultad"`
	Respuesta     string    `json:"respuesta"`
	Correcta      bool      `json:"correcta"`
	DuracionMS    int64     `json:"duracionMs"`
	Intento       int       `json:"intento"`
	Ayuda         string    `json:"ayuda"`
	Momento       time.Time `json:"momento"`
}

type ReglaAdaptativa struct {
	ID          string `json:"id"`
	Descripcion string `json:"descripcion"`
}

type DecisionAdaptativa struct {
	ReglaID            string  `json:"reglaId"`
	Razon              string  `json:"razon"`
	TemaSiguiente      string  `json:"temaSiguiente"`
	NivelSiguiente     int     `json:"nivelSiguiente"`
	PrecisionReciente  float64 `json:"precisionReciente"`
	ErroresConsecutivos int    `json:"erroresConsecutivos"`
	Dominio            float64 `json:"dominio"`
	InteraccionID      string  `json:"interaccionId"`
}

type Actividad struct {
	ID          string   `json:"id"`
	Tema        string   `json:"tema"`
	Nivel       int      `json:"nivel"`
	Enunciado   string   `json:"enunciado"`
	Opciones    []string `json:"opciones"`
	Respuesta   string   `json:"-"`
	Explicacion string   `json:"-"`
}

var Reglas = []ReglaAdaptativa{
	{"RA01", "Diagnóstico establece el nivel inicial"},
	{"RA02", "Nueva interacción actualiza el perfil y dominio"},
	{"RA03", "Primer error ofrece corrección y pista"},
	{"RA04", "Error recurrente cambia apoyo y reduce dificultad"},
	{"RA05", "Precisión alta y respuesta rápida aumenta dificultad"},
	{"RA06", "Dominio insuficiente programa refuerzo"},
	{"RA07", "Dominio alto permite avanzar de tema"},
}

var Actividades = []Actividad{
	{"v1a", "vocabulario", 1, "I drink ___ in the morning.", []string{"water", "chair", "blue"}, "water", "Water significa agua y completa la oración."},
	{"v1b", "vocabulario", 1, "A ___ is a place to read books.", []string{"library", "kitchen", "shoe"}, "library", "Library significa biblioteca."},
	{"v2a", "vocabulario", 2, "Choose the word closest to 'quick'.", []string{"fast", "late", "quiet"}, "fast", "Quick y fast significan rápido."},
	{"v2b", "vocabulario", 2, "She bought groceries at the ___.", []string{"market", "station", "hospital"}, "market", "Market es un lugar donde se compran alimentos."},
	{"v3a", "vocabulario", 3, "Choose the best meaning of 'reliable'.", []string{"dependable", "temporary", "distant"}, "dependable", "Reliable significa confiable; dependable es equivalente."},
	{"v3b", "vocabulario", 3, "The evidence was ___; we could trust it.", []string{"credible", "fragile", "vague"}, "credible", "Credible significa creíble."},
	{"g1a", "gramatica", 1, "She ___ a student.", []string{"is", "are", "am"}, "is", "She usa is en presente del verbo to be."},
	{"g1b", "gramatica", 1, "They ___ at home.", []string{"are", "is", "am"}, "are", "They usa are en presente del verbo to be."},
	{"g2a", "gramatica", 2, "He ___ to class every day.", []string{"goes", "go", "going"}, "goes", "En presente simple, he requiere la forma goes."},
	{"g2b", "gramatica", 2, "We ___ English on Mondays.", []string{"study", "studies", "studying"}, "study", "We usa la forma base study en presente simple."},
	{"g3a", "gramatica", 3, "If it rains, we ___ home.", []string{"will stay", "stayed", "staying"}, "will stay", "El primer condicional usa will en la consecuencia."},
	{"g3b", "gramatica", 3, "She has lived here ___ 2020.", []string{"since", "for", "during"}, "since", "Since introduce el punto inicial de un periodo."},
}

func BuscarActividad(id string) (Actividad, bool) {
	for _, a := range Actividades { if a.ID == id { return a, true } }
	return Actividad{}, false
}
