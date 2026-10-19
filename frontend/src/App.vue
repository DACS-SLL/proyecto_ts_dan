<script setup>
import { ref, computed, onMounted } from 'vue'

async function api(path, options = {}) {
  const response = await fetch('/api' + path, { headers: { 'Content-Type': 'application/json' }, ...options })
  const data = await response.json()
  if (!response.ok) throw new Error(data.error || 'No se pudo completar la solicitud')
  return data
}

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
</script>

<template>
<main class="shell">
  <header class="page-header">
    <p class="version">Prototipo de tesis · v0.1</p>
    <h1>Aprendizaje adaptativo de inglés</h1>
    <p>Diagnóstico, práctica y actividades que se ajustan a tus respuestas.</p>
  </header>
  <p v-if="error" class="error" role="alert">{{ error }}</p>
  <section v-if="!sesion" class="panel diagnosis">
    <h2>Diagnóstico inicial</h2>
    <p>Responde estas cuatro preguntas para comenzar con actividades de vocabulario y gramática.</p>
    <label class="name">Tu nombre <input v-model="nombre" maxlength="80" placeholder="Escribe tu nombre"></label>
    <fieldset v-for="(p, i) in preguntas" :key="p.id"><legend>{{ i + 1 }}. {{ p.enunciado }} <span>{{ temaNombre(p.tema) }}</span></legend>
      <label v-for="opcion in p.opciones" :key="opcion" class="option"><input type="radio" :name="p.id" :value="opcion" v-model="diagnostico[p.id]">{{ opcion }}</label>
    </fieldset>
    <button :disabled="!listo || cargando" @click="crear">Comenzar actividades</button>
  </section>
  <template v-else>
    <section class="profile-line"><div><h2>{{ sesion.perfil.nombre }}</h2><p>Tu sesión de práctica</p></div><button class="secondary" @click="reiniciar">Empezar de nuevo</button></section>
    <section class="grid">
      <div class="panel practice"><h2>{{ temaNombre(sesion.actividad.tema) }}</h2><p class="level">Nivel {{ sesion.actividad.nivel }}</p><p class="question">{{ sesion.actividad.enunciado }}</p>
        <template v-if="sesion.pendiente"><div class="answer-list"><label v-for="opcion in sesion.actividad.opciones" :key="opcion" class="option"><input type="radio" name="respuesta" :value="opcion" v-model="respuesta">{{ opcion }}</label></div><button :disabled="!respuesta || cargando" @click="responder">Comprobar respuesta</button></template>
        <template v-else><div class="feedback" :class="sesion.ultima.correcta ? 'success' : 'needs-work'"><strong>{{ sesion.ultima.correcta ? 'Correcto' : 'Revisemos la respuesta' }}</strong><p>{{ sesion.feedback }}</p></div><button :disabled="cargando" @click="siguiente">Siguiente actividad</button></template>
      </div>
      <aside class="panel progress"><h2>Tu progreso</h2><p class="tally">{{ sesion.progreso.aciertos }} de {{ sesion.progreso.respondidas }} actividades correctas</p><div v-for="estado in sesion.perfil.temas" :key="estado.tema" class="topic"><div class="topic-head"><strong>{{ temaNombre(estado.tema) }}</strong><span>{{ porcentaje(estado.dominio) }}</span></div><div class="bar"><i :style="{width: porcentaje(estado.dominio)}"></i></div><small>Nivel {{ estado.nivel }} · {{ estado.evidencias }} respuestas consideradas{{ estado.refuerzo ? ' · necesita repaso' : '' }}</small></div><p class="note">El porcentaje es una estimación basada en tus respuestas.</p></aside>
    </section>
    <section v-if="sesion.decision" class="panel decision"><h2>Cómo se ajustó la actividad</h2><p>{{ sesion.decision.razon }}</p><dl><div><dt>Precisión reciente</dt><dd>{{ porcentaje(sesion.decision.precisionReciente) }}</dd></div><div><dt>Errores seguidos</dt><dd>{{ sesion.decision.erroresConsecutivos }}</dd></div><div><dt>Ruta sugerida</dt><dd>{{ temaNombre(sesion.decision.temaSiguiente) }} · nivel {{ sesion.decision.nivelSiguiente }}</dd></div></dl><small>Regla {{ sesion.decision.reglaId }}</small></section>
  </template>
</main>
</template>
