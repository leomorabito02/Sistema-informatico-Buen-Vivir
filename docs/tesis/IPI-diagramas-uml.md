# Diagramas UML para el IPI (Figuras 35 y 36)

Código Mermaid verificado contra el código de `education/` (modelos y
`CreateWithAttendance` en `submission_repository.go`) y los consumidores de
`dashboard/` y `human-promotion/`.

Para obtener la imagen: pegar cada bloque en https://mermaid.live, exportar PNG e
insertarlo en el documento, en la sección "Comunicación entre módulos" (después de
la Figura 34), con número, título en cursiva, descripción y fuente.

## Figura 35. Diagrama de clases del dominio de Educación

Descripción sugerida: Entidades del servicio de Educación y sus relaciones. La
entrega de una actividad genera la asistencia y otorga las habilidades asociadas a
esa actividad. Fuente: elaboración propia.

```mermaid
classDiagram
direction LR
class Curso {
  id
  titulo
  modalidad
  cupo
  areaFormativa
}
class Actividad {
  id
  cursoId
  titulo
  consigna
  fecha
}
class Inscripcion {
  id
  participanteId
  cursoId
  fechaInscripcion
  estado
}
class Entrega {
  id
  participanteId
  actividadId
  texto
  archivos
  fechaEntrega
}
class Asistencia {
  id
  participanteId
  actividadId
  presente
  fecha
}
class Habilidad {
  id
  nombre
  categoria
}
class ActividadHabilidad {
  actividadId
  habilidadId
  cursoId
}
class HabilidadParticipante {
  participanteId
  habilidadId
  origen
  fecha
}
Curso "1" --> "0..*" Actividad
Curso "1" --> "0..*" Inscripcion
Actividad "1" --> "0..*" Entrega
Actividad "1" --> "0..*" Asistencia
Actividad "1" --> "0..*" ActividadHabilidad
Habilidad "1" --> "0..*" ActividadHabilidad
Habilidad "1" --> "0..*" HabilidadParticipante
Entrega ..> Asistencia : genera
Entrega ..> HabilidadParticipante : otorga
```

## Figura 36. Diagrama de secuencia de la entrega de una actividad

Descripción sugerida: Flujo síncrono (REST) y asíncrono (Pub/Sub con Transactional
Outbox) desde que un participante entrega una actividad hasta que Dashboard y
Promoción Humana reciben los eventos. Fuente: elaboración propia.

```mermaid
sequenceDiagram
autonumber
actor P as Participante
participant F as Frontend PWA
participant G as API Gateway
participant E as Educación
participant S as Cloud Storage
participant D as Firestore
participant B as Pub/Sub
participant H as Promoción Humana
participant R as Dashboard
P->>F: Completa la entrega
F->>G: POST /actividades/{id}/entrega (JWT)
G->>E: Reenvía la solicitud
E->>D: Verifica actividad e inscripción activa
E->>S: Sube los adjuntos
E->>D: Transacción: entrega, asistencia, habilidades y outbox_eventos
E-->>F: 201 Created
E->>B: Relay publica los eventos del outbox
B->>R: Push asistencia y curso completado
R->>R: Invalida la caché del resumen
B->>H: Push curso completado (última actividad)
H->>H: Registra el hito en la línea de tiempo
```
