import assert from 'node:assert/strict'
import test from 'node:test'
import { diagnosticoCompleto, resumenDiagnostico } from './diagnostic.js'

const preguntas = [
  { id: 'd1', opciones: ['manzana', 'mesa'] },
  { id: 'd2', opciones: ['casa', 'perro'] }
]

test('solo permite iniciar con nombre y respuestas válidas para todas las preguntas', () => {
  assert.equal(diagnosticoCompleto('Ana', preguntas, { d1: 'manzana', d2: 'casa' }), true)
  assert.equal(diagnosticoCompleto('  ', preguntas, { d1: 'manzana', d2: 'casa' }), false)
  assert.equal(diagnosticoCompleto('Ana', preguntas, { d1: 'manzana' }), false)
  assert.equal(diagnosticoCompleto('Ana', preguntas, { d1: 'desconocida', d2: 'casa' }), false)
  assert.equal(diagnosticoCompleto('Ana', [], {}), false)
})

test('muestra los resultados iniciales por tema en orden estable', () => {
  const perfil = { diagnostico: {
    gramatica: { tema: 'gramatica', aciertos: 0, total: 2 },
    vocabulario: { tema: 'vocabulario', aciertos: 1, total: 2 }
  } }
  assert.deepEqual(resumenDiagnostico(perfil).map(r => [r.tema, r.aciertos, r.total]), [
    ['vocabulario', 1, 2], ['gramatica', 0, 2]
  ])
  assert.deepEqual(resumenDiagnostico({}), [])
})
