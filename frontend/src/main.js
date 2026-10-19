import { createApp, ref, computed, onMounted } from 'vue'
import './style.css'

async function api(path, options = {}) {
  const response = await fetch('/api' + path, { headers: { 'Content-Type': 'application/json' }, ...options })
  const data = await response.json()
  if (!response.ok) throw new Error(data.error || 'No se pudo completar la solicitud')
  return data
}

createApp({
  setup() {
    const preguntas = ref([])
    const nombre = ref('')
    const diagnostico = ref({})
    const sesion = ref(null)
    const respuesta = ref('')
    const inicio = ref(Date.now())
    const cargando = ref(false)
    const error = ref('')
    const temaNombre = tema => tema === 'gramatica' ? 'Gramática' : 'Vocabulario'
    const porcentaje = valor => Math.round(valor * 100) + '%'
    const listo = computed(() => preguntas.value.length > 0 && nombre.value.trim() && preguntas.value.every(p => diagnostico.value[p.id]))
    const resetActividad = () => { respuesta.value = ''; inicio.value = Date.now() }
    const ejecutar = async trabajo => {
      error.value = ''; cargando.value = true
      try { await trabajo() } catch (e) { error.value = e.message } finally { cargando.value = false }
    }
    onMounted(() => ejecutar(async () => {
      preguntas.value = await api('/diagnostic')
      const id = localStorage.getItem('tesis-v01-sesion')
      if (id) {
        try { sesion.value = await api('/sessions/' + id) } catch { localStorage.removeItem('tesis-v01-sesion') }
      }
    }))
    const crear = () => ejecutar(async () => {
      sesion.value = await api('/sessions', { method: 'POST', body: JSON.stringify({ nombre: nombre.value, respuestas: diagnostico.value }) })
      localStorage.setItem('tesis-v01-sesion', sesion.value.perfil.id)
      resetActividad()
    })
    const responder = () => ejecutar(async () => {
      sesion.value = await api('/sessions/' + sesion.value.perfil.id + '/answer', {
        method: 'POST', body: JSON.stringify({ actividadId: sesion.value.actividad.id, respuesta: respuesta.value, duracionMs: Date.now() - inicio.value })
      })
    })
    const siguiente = () => ejecutar(async () => {
      sesion.value = await api('/sessions/' + sesion.value.perfil.id + '/next', { method: 'POST' })
      resetActividad()
    })
    const reiniciar = () => { localStorage.removeItem('tesis-v01-sesion'); sesion.value = null; diagnostico.value = {}; nombre.value = ''; error.value = '' }
    return { preguntas, nombre, diagnostico, sesion, respuesta, cargando, error, listo, temaNombre, porcentaje, crear, responder, siguiente, reiniciar }
  },
  template: `
    <main class="shell">
      <header><div><small>TESIS · DEMOSTRACIÓN PRELIMINAR</small><h1>Aprendizaje adaptativo de inglés</h1><p>Prototipo v0.1: diagnóstico, práctica y reglas explicables.</p></div><span class="badge">Go + Vue</span></header>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <section v-if="!sesion" class="card">
        <div class="eyebrow">Paso 1 · Diagnóstico inicial</div>
        <h2>Conozcamos tu punto de partida</h2>
        <p>Cuatro preguntas breves estiman un nivel inicial de vocabulario y gramática. No son una prueba validada.</p>
        <label class="name">Nombre del estudiante <input v-model="nombre" maxlength="80" placeholder="Escribe tu nombre"></label>
        <fieldset v-for="(p, i) in preguntas" :key="p.id"><legend>{{ i + 1 }}. {{ p.enunciado }} <span>{{ temaNombre(p.tema) }}</span></legend>
          <label v-for="opcion in p.opciones" :key="opcion" class="option"><input type="radio" :name="p.id" :value="opcion" v-model="diagnostico[p.id]">{{ opcion }}</label>
        </fieldset>
        <button :disabled="!listo || cargando" @click="crear">Crear perfil y empezar</button>
      </section>
      <template v-else>
        <section class="summary card"><div><div class="eyebrow">Perfil del estudiante</div><h2>{{ sesion.perfil.nombre }}</h2><p>Versión {{ sesion.perfil.version }} · Actualizado {{ new Date(sesion.perfil.actualizadoEn).toLocaleString() }}</p></div><button class="secondary" @click="reiniciar">Nueva demostración</button></section>
        <section class="grid">
          <div class="card practice"><div class="eyebrow">Paso 2 · Actividad</div><h2>{{ temaNombre(sesion.actividad.tema) }} · nivel {{ sesion.actividad.nivel }}</h2><p class="question">{{ sesion.actividad.enunciado }}</p>
            <template v-if="sesion.pendiente"><label v-for="opcion in sesion.actividad.opciones" :key="opcion" class="option"><input type="radio" name="respuesta" :value="opcion" v-model="respuesta">{{ opcion }}</label><button :disabled="!respuesta || cargando" @click="responder">Evaluar respuesta</button></template>
            <template v-else><div class="feedback" :class="sesion.ultima.correcta ? 'success' : 'needs-work'"><strong>{{ sesion.ultima.correcta ? 'Acierto' : 'Para revisar' }}</strong><p>{{ sesion.feedback }}</p><small>Apoyo registrado: {{ sesion.ultima.ayuda || 'confirmación' }}</small></div><button :disabled="cargando" @click="siguiente">Ir a la siguiente actividad</button></template>
          </div>
          <aside class="card"><div class="eyebrow">Paso 3 · Progreso básico</div><h2>Por tema</h2><div v-for="estado in sesion.perfil.temas" :key="estado.tema" class="topic"><div class="topic-head"><strong>{{ temaNombre(estado.tema) }}</strong><span>{{ porcentaje(estado.dominio) }} estimado</span></div><div class="bar"><i :style="{width: porcentaje(estado.dominio)}"></i></div><small>Nivel {{ estado.nivel }} · {{ estado.evidencias }} evidencias{{ estado.refuerzo ? ' · refuerzo pendiente' : '' }}</small></div><p class="tally">{{ sesion.progreso.aciertos }} aciertos de {{ sesion.progreso.respondidas }} actividades</p><p class="note">El dominio es una estimación didáctica, no una medición validada.</p></aside>
        </section>
        <section v-if="sesion.decision" class="card decision"><div class="eyebrow">Decisión adaptativa trazable</div><h2>{{ sesion.decision.reglaId }}</h2><p>{{ sesion.decision.razon }}</p><div class="metrics"><span>Precisión reciente: {{ porcentaje(sesion.decision.precisionReciente) }}</span><span>Errores seguidos: {{ sesion.decision.erroresConsecutivos }}</span><span>Siguiente: {{ temaNombre(sesion.decision.temaSiguiente) }} · nivel {{ sesion.decision.nivelSiguiente }}</span></div></section>
      </template>
      <footer>La retroalimentación v0.1 usa textos predefinidos; la IA queda como extensión futura.</footer>
    </main>`
}).mount('#app')
