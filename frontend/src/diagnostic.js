export function diagnosticoCompleto(nombre, preguntas, respuestas) {
  return Boolean(nombre.trim()) && preguntas.length > 0 && preguntas.every(p =>
    p.opciones.includes(respuestas[p.id])
  )
}

export function resumenDiagnostico(perfil) {
  return ['vocabulario', 'gramatica']
    .map(tema => perfil.diagnostico?.[tema])
    .filter(Boolean)
}
