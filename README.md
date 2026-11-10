# Artefacto adaptativo v0.1

Prototipo web **preliminar** de la tesis de Daniel. Demuestra el núcleo adaptativo con Go y una interfaz sencilla en Vue. No es una aplicación final ni una evaluación pedagógica validada. La lógica del backend está separada de la interfaz para servir de base a una versión móvil posterior.

## Alcance demostrado

1. Cuatro preguntas diagnósticas crean el perfil, conservan el resultado inicial por tema y estiman el punto de partida de vocabulario y gramática.
2. Una actividad de opción múltiple produce una interacción con resultado, tiempo, intento, tema, nivel y ayuda.
3. El dominio por tema se recalcula con la estimación previa y la precisión reciente. Cada respuesta incrementa la versión del perfil sin alterar el resultado diagnóstico inicial.
4. Reglas explícitas deciden mantener, subir o bajar el nivel, cambiar el apoyo ante errores repetidos y marcar refuerzo.
5. La interfaz muestra feedback, razón de la decisión, siguiente actividad y progreso básico.

Los umbrales se definen en `backend/domain/model.go` (`ConfigUmbrales`): dos errores recurrentes, dominio alto 0,75, refuerzo por debajo de 0,5, respuesta rápida hasta 12 segundos y ventana de cuatro interacciones. Son parámetros **ilustrativos** para la demostración; requieren validación docente y empírica.

## Arquitectura

```text
Vue (diagnóstico y práctica)
        | JSON HTTP
Go api (sesiones en memoria y validación)
        |
Go domain (PerfilEstudiante, EstadoTema, Interaccion,
           ReglaAdaptativa, DecisionAdaptativa, motor y catálogo)
```

El módulo `domain` no conoce la interfaz web. Las respuestas correctas permanecen en el servidor. El endpoint de interacciones permite inspeccionar la evidencia de la sesión.

## Cómo ejecutar

Requiere Go 1.22 o posterior, Node.js y npm. En dos terminales:

```sh
cd backend
go run .
```

```sh
cd frontend
npm install
npm run dev
```

Abrir la URL local que muestra Vite (normalmente `http://127.0.0.1:5173`). La API escucha solo en `127.0.0.1:8080`; Vite reenvía `/api` a ese puerto.

Para verificar: `cd backend && go test ./...`; `cd frontend && npm test && npm run build`.

### API mínima

| Método | Ruta | Uso |
|---|---|---|
| GET | `/api/diagnostic` | Preguntas sin respuestas correctas |
| POST | `/api/sessions` | Crear perfil con nombre y mapa de respuestas `d1`–`d4` |
| GET | `/api/sessions/{id}` | Estado, actividad y progreso |
| POST | `/api/sessions/{id}/answer` | Evaluar `{actividadId, respuesta, duracionMs}` |
| POST | `/api/sessions/{id}/next` | Presentar actividad elegida |
| GET | `/api/sessions/{id}/interactions` | Historial de evidencias |

## Trazabilidad con la especificación

| Requisito | Demostración v0.1 | Archivo principal |
|---|---|---|
| ESP-0001 | Diagnóstico inicial con aciertos, total y estimación inicial por tema | `backend/domain/engine.go` |
| ESP-0002 | Perfil versionado, resultado inicial conservado y dominio actualizado tras cada respuesta | `backend/domain/model.go`, `backend/domain/engine.go` |
| ESP-0003 | Interacción recuperable durante la sesión | `backend/api/server.go` |
| ESP-0005 | Dominio por tema y precisión reciente | `backend/domain/engine.go` |
| ESP-0006 | RA04 cambia el apoyo ante error recurrente; RA05 reduce el nivel ante dificultad elevada; RA06 aumenta el nivel o avanza de tema por dominio alto | `backend/domain/engine.go` |
| ESP-0008 | Corrección y explicación de la respuesta | `backend/domain/engine.go` |
| ESP-0009 | RA04 cambia a apoyo alternativo por error recurrente | `backend/domain/engine.go` |
| ESP-0010 | RA07 mantiene el tema para refuerzo en la ruta inmediata | `backend/domain/engine.go` |
| ESP-0014 | Resultado inicial, aciertos en práctica y dominio estimado por tema | `frontend/src/App.vue` |

ESP-0015 corresponde al **panel de seguimiento docente**. No existe todavía en este prototipo; la configuración interna de umbrales no implementa ese requisito.

### Correspondencia con los diagramas de Fase 3

La matriz de trazabilidad de la tesis vincula este primer incremento especialmente con A1 (clases del dominio), A2 (componentes), A3 (casos de uso), B1 (capas), B3 (secuencia del diagnóstico), B6 (navegación) y C1 (estados de sesión). El flujo ya existente de práctica se relaciona con A4 (IDEF3), B4 (secuencia de respuesta) y B5 (motor). B2 (modelo lógico de datos) y C2 (despliegue) representan diseño previsto: la v0.1 mantiene sesiones en memoria y no establece persistencia definitiva ni un despliegue de producción. Las figuras originales A1–A4, B1–B6 y C1–C2 no están incluidas en esta copia local; esta correspondencia utiliza la matriz acordada en la revisión de tesis y requiere cotejo visual con los archivos de los diagramas.

## Límites de esta versión

- Las sesiones viven en memoria: al reiniciar Go se pierden perfiles e interacciones. No hay cuentas ni autenticación.
- El diagnóstico, el catálogo y las fórmulas son pequeños e ilustrativos. El resultado inicial se conserva por tema, pero no registra todavía cada respuesta diagnóstica como interacción individual. La estimación de dominio no debe interpretarse como medición validada.
- El refuerzo reaparece en la ruta de práctica inmediata; la programación entre sesiones queda para una siguiente versión.
- La retroalimentación usa textos predefinidos. Una interfaz de IA para feedback puede incorporarse más adelante; ninguna decisión adaptativa depende de ML.
- No hay panel docente, análisis avanzado ni aplicación móvil. La migración móvil utilizaría la API o la lógica de dominio, según la arquitectura final.
