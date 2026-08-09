HISTORIAS DE USUARIO
1. Autenticación y Gestión de Usuarios

HU-01 (RF-01)
Prioridad: Alta
Como administrador del sistema, quiero registrar usuarios con nombre, email, contraseña, teléfono y rol, para dar de alta a los distintos perfiles (participante, voluntario, empresa, docente) en la plataforma.

Criterios de aceptación:
Dado que soy administrador y completo nombre, email, contraseña, teléfono y un rol válido, cuando guardo el formulario, entonces el usuario se crea correctamente.
El sistema debe validar que el email no esté duplicado.
La contraseña debe cumplir una política mínima de seguridad (8 caracteres. Al menos una mayúscula un caracter especial).
Si falta un campo obligatorio, el sistema debe mostrar un error y no crear el usuario.
El email debe tener dominio gmail

HU-02 (RF-02)
Prioridad: Alta
Como usuario registrado, quiero iniciar sesión con mi email/nombre de usuario y contraseña, para acceder a las funcionalidades de la app según mi rol.

Criterios de aceptación:
Dado credenciales válidas, cuando envío el formulario de login, entonces recibo un JWT válido y soy redirigido a mi panel correspondiente.
Dado credenciales inválidas, cuando intento iniciar sesión, entonces el sistema muestra un error genérico sin indicar cuál dato es incorrecto.
El sistema debe bloquear temporalmente la cuenta tras 5 intentos fallidos consecutivos.

HU-02bis
Prioridad: Alta

Como usuario registrado, quiero recuperar mi contraseña mediante un enlace enviado a mi email, para poder acceder a mi cuenta si la olvidé.

Criterios de aceptación:
Dado que solicito "olvidé mi contraseña" e ingreso mi email, cuando el email existe en el sistema, entonces recibo un enlace de recuperación por correo.
El enlace de recuperación debe expirar a los 30 minutos de generado.
El enlace debe ser de un solo uso.
Dado un enlace válido, cuando ingreso una nueva contraseña que cumple la política de seguridad, entonces la contraseña se actualiza y se invalidan las sesiones activas previas.
Si el email no existe en el sistema, el mensaje debe ser genérico (no debe confirmar ni desmentir si el email está registrado), para evitar filtrar información de usuarios.
HU-03 (RF-03)
Prioridad: Alta
Como sistema, quiero validar la autenticidad, integridad y vigencia del JWT en cada solicitud a un recurso protegido, para garantizar que solo usuarios autenticados accedan a la información.

Criterios de aceptación:
Dado un JWT válido y vigente, cuando se solicita un recurso protegido, entonces la solicitud se procesa normalmente.
Dado un JWT expirado, alterado o inválido, cuando se solicita un recurso protegido, entonces el sistema responde 401 y deniega el acceso.
Toda validación fallida debe quedar registrada en el log de auditoría.

HU-04 (RF-04)
Prioridad: Alta
Como responsable de seguridad del sistema, quiero que los JWT tengan expiración configurable y los refresh tokens sean de uso único, para reducir el riesgo de reutilización de tokens.

Criterios de aceptación:
El tiempo de expiración del JWT debe ser parametrizable sin modificar código.
Dado un refresh token ya utilizado, cuando se intenta reutilizar, entonces el sistema lo rechaza e invalida la sesión asociada.
Cada rotación de refresh token debe invalidar el token anterior.

HU-05 (RF-05)
Prioridad: Alta
Como usuario del sistema, quiero contar con autenticación de doble factor (2FA), para proteger mi cuenta ante accesos no autorizados.

Criterios de aceptación:
Dado que tengo 2FA habilitado, cuando inicio sesión con credenciales correctas, entonces el sistema solicita un segundo código antes de otorgar acceso.
Si el código 2FA es incorrecto o expiró, el acceso debe ser denegado.
El sistema debe permitir activar/desactivar el 2FA desde el perfil del usuario.

HU-06 (RF-06)
Prioridad: Alta
Como administrador, quiero que el control de acceso siga un modelo RBAC con permisos granulares por rol, para asegurar que cada usuario solo acceda a los recursos correspondientes a su función.

Criterios de aceptación:
Dado un usuario con un rol determinado, cuando intenta acceder a un recurso fuera de su alcance, entonces el sistema deniega el acceso (403).
Los permisos por rol deben poder configurarse y consultarse de forma centralizada en el panel de configuración del administrador del sistema.
Todo intento de acceso no autorizado debe quedar registrado en el log de auditoría.
Cada usuario según su rol, debe tener acceso a las diferentes vistas: 
Vista administrador:
Gestión de usuarios: crear/editar/eliminar cualquier perfil, aprobar voluntarios y empresas
Gestión de cursos (crear/editar/eliminar), moderación de contenido subido por docentes externos
Fichas de proceso personal: crear, editar, ver línea de tiempo, exportar PDF
Gestión de ofertas laborales: aprobar, rechazar, eliminar
Gestión de campañas: crear, editar, publicar
Dashboard completo con filtros y exportación de reportes
Log de auditoría completo (lectura, filtrado)
Panel de configuración RBAC (permisos por rol)
Vista voluntario:
Crear/editar/eliminar perfiles de participantes (con aprobación previa de su propia cuenta)
Inscribir participantes a cursos, registrar asistencia
Fichas de proceso personal: crear, editar, registrar hitos, ver línea de tiempo, exportar PDF, buscar participantes
Crear/editar/publicar campañas
Consultar catálogo de cursos y ofertas laborales (sin poder aprobarlas)
Sin acceso al dashboard analítico completo ni al panel RBAC ni al log de auditoría global
Vista participante:
Ver su propio perfil y ficha de proceso personal (datos no confidenciales)
Ver su línea de tiempo de actividad
Ver progreso formativo en cursos en los que está inscripto
Acceder al contenido educativo del centro (consulta de materiales de cursos y biblioteca)
Ver ofertas laborales sugeridas y generar su CV
Buscar cursos, ofertas laborales y campañas
Comentar contenido público
Configurar sus notificaciones
Vista empresa:
Publicar ofertas laborales (tras aprobación de su cuenta)
Ver el estado de sus propias ofertas (pendiente/publicada/rechazada)
Ver postulantes de sus ofertas
Buscar y ver campañas y contenido público
Sin acceso a perfiles de participantes ni fichas de proceso personal
Vista docente (responsable de curso):
Registro con aprobación previa del coordinador/administrador, con el mismo flujo que el voluntario
Consultar los cursos donde figura como docente responsable
Editar el contenido y materiales de sus propios cursos (adjuntar/actualizar archivos)
Registrar asistencia de participantes mediante el cumplimiento de actividades
Consultar el progreso formativo y el listado de participantes inscriptos en sus cursos
Consultar contenido educativo del centro
Sin acceso a: crear/eliminar cursos, perfiles de participantes, fichas de proceso personal, dashboard, gestión de usuarios ni moderación de contenido de docentes externos
Vista docente externo:
Subir contenido educativo propio (queda en estado "pendiente de moderación")
Consultar el estado de sus materiales subidos (pendiente/aprobado/rechazado)
Consultar contenido educativo público habilitado del centro
Editar/eliminar sus propios materiales mientras estén en estado "pendiente"
Sin acceso a perfiles de participantes, fichas de proceso personal, dashboard, creación de cursos ni gestión de usuarios
HU-07 (RF-07)
Prioridad: Alta
Como administrador o voluntario, quiero crear perfiles de participantes, para registrar a las personas que ingresan al Centro de Ecología Integral.

Criterios de aceptación:
Dado que soy administrador o voluntario y completo los datos obligatorios del participante, cuando guardo, entonces el perfil se crea correctamente.
El sistema debe validar los campos obligatorios (nombre, apellido, email, dni, teléfono, dirección y fecha de nacimiento) antes de guardar.
La creación debe quedar registrada en el log de auditoría con usuario y fecha.

HU-08 (RF-08)
Prioridad: Alta
Como administrador o voluntario, quiero editar perfiles de participantes existentes, para mantener actualizada su información.

Criterios de aceptación:
Dado un perfil existente, cuando modifico uno o más campos y guardo, entonces los cambios se reflejan inmediatamente y aparece un mensaje de exito: “Cambios guardados correctamente”
La edición debe quedar registrada en el log de auditoría con el detalle de los campos modificados.

HU-09 (RF-09)
Prioridad: Alta
Como administrador o voluntario, quiero eliminar perfiles de participantes, para dar de baja registros que ya no corresponde mantener activos.

Criterios de aceptación:
Dado un perfil existente, cuando confirmó la eliminación, entonces el perfil dejá de estar disponible en el sistema (baja lógica).
Cuando se confirma la eliminación, el perfil de participante debe dejar de estar visible en las listas de participantes a las que tienen acceso el voluntario y administrador.
El sistema debe solicitar una confirmación explícita antes de eliminar.
La eliminación debe quedar registrada en el log de auditoría.

HU-10 (RF-10)
Prioridad: Alta
Como administrador, quiero que el sistema registre en un log de auditoría todas las acciones críticas sobre perfiles de participantes, para garantizar trazabilidad y cumplimiento normativo.

Criterios de aceptación:
Dado que se ejecuta una acción crítica (alta, baja, modificación) sobre un perfil, cuando la operación se completa, entonces se genera un registro con usuario, acción, recurso y fecha.
El log de auditoría debe ser inmutable y de solo lectura para los usuarios.
El administrador debe poder consultar el log filtrando por usuario, fecha o tipo de acción.


HU-11 (RF-11)
Prioridad: Media

Como voluntario, quiero editar el perfil de un participante incluyendo datos personales, foto y documentación, para mantener su información completa y actualizada.

Criterios de aceptación:
Dado un perfil de participante, cuando adjunto o reemplazo una foto o documento, entonces el archivo se almacena y se asocia correctamente al perfil.
El sistema debe validar formato y tamaño máximo de los archivos adjuntos.
Foto de perfil: formatos JPG, JPEG, PNG o WEBP; tamaño máximo 5 MB.
Documentación (DNI, certificados, etc.): formatos PDF, JPG, JPEG o PNG; tamaño máximo 10 MB por archivo.
Los cambios deben quedar registrados en el log de auditoría.

HU-12 (RF-12)
Prioridad: Alta
Como coordinador del centro barrial, quiero aprobar el registro de nuevos voluntarios,
para garantizar que solo personas validadas por el centro puedan operar en la plataforma.

Criterios de aceptación:
Dado un voluntario recién registrado, cuando el administrador revisa la solicitud, entonces puede aprobarla o rechazarla.
Mientras la solicitud está pendiente, el voluntario no debe tener acceso a funcionalidades del sistema.
El voluntario debe recibir una notificación con el resultado de la aprobación/rechazo.

HU-13 (RF-13)
Prioridad: Media
Como usuario, quiero autenticarse utilizando mi cuenta de Google, para acceder al sistema de forma rápida sin crear una contraseña nueva.

Criterios de aceptación:
Dado que selecciono 'Iniciar sesión con Google', cuando autorizó el acceso, entonces el sistema crea o vincula mi cuenta y me autentica.
El sistema debe guardar el identificador de Google (googleAuth) asociado al usuario.
Si el email de Google ya existe en el sistema, la cuenta debe vincularse en lugar de duplicarse.

2. Educación

HU-14 (RF-14)
Prioridad: Alta
Como administrador, quiero crear cursos con título, descripción, área formativa, docente responsable, para ofrecer nueva formación a los participantes.

Criterios de aceptación:
Dado que completo los campos obligatorios del curso, cuando guardo, entonces el curso queda creado y visible en el catálogo.
El sistema debe permitir adjuntar materiales (archivos/enlaces) al curso.
Debe validarse que exista un docente responsable asignado.

HU-15 (RF-15)
Prioridad: Alta
Como administrador, quiero editar cursos existentes (título, descripción, área, docente), para mantener actualizada la oferta formativa.

Criterios de aceptación:
Dado un curso existente, cuando modificó sus datos y guardó, entonces los cambios se reflejan en el catálogo y aparece un mensaje de éxito “Se modificó el curso correctamente”
Los participantes ya inscritos deben poder ver la versión actualizada del curso.
La actualización de los cursos deben verse reflejadas en la base de datos

HU-16 (RF-16)
Prioridad: Alta
Como administrador, quiero eliminar cursos, para retirar de la oferta formativa aquellos que ya no se dictan.

Criterios de aceptación:
Dado un curso existente, cuando confirmo su eliminación, entonces el curso pasa a estado "inactivo/dado de baja" (baja lógica), deja de ser visible en el catálogo público, pero se conserva en la base de datos, y aparece un mensaje de éxito "Se eliminó el curso correctamente".
El sistema debe advertir si existen participantes inscriptos antes de eliminar, mostrando la cantidad de inscriptos activos.
Los participantes que ya cursaron o están cursando deben seguir viendo el curso en su propio historial, aunque ya no esté visible en el catálogo general.
La baja debe quedar registrada en el log de auditoría.

HU-17 (RF-17)
Prioridad: Alta
Como administrador,quiero gestionar el contenido de los cursos a través de Contentful (CMS headless),para centralizar y facilitar la edición de contenido educativo.

Criterios de aceptación:
Dado que edito contenido en Contentful, cuando publico los cambios, entonces se reflejan en la plataforma.
Si Contentful no responde (error 5xx o expiración), el sistema debe mostrar el último contenido cacheado (stale content) sin romper la interfaz.

HU-18 (RF-18)
Prioridad: Alta
Como voluntario o responsable, quiero inscribir participantes a cursos desde el catálogo,
para gestionar su ingreso a instancias formativas.

Criterios de aceptación:
Dado un participante y un curso con cupo disponible, cuando confirmo la inscripción, entonces el participante queda inscripto y aparece un mensaje de exito “Se inscribe al participante correctamente”
El sistema debe impedir inscripciones duplicadas al mismo curso y mostrar mensaje de error: “El participante ya se encuentra inscripto en el curso”
El participante debe recibir una notificación de confirmación por email.

HU-19 (RF-19)
Prioridad: Alta
Como voluntario o docente, quiero registrar la asistencia de los participantes mediante el cumplimiento de actividades, para llevar un control del avance formativo.

Criterios de aceptación:
Dado un participante inscripto en un curso, cuando marco una actividad como completada, entonces se actualiza su registro de asistencia.
El sistema debe permitir consultar el historial de asistencia por participante y por curso.

HU-19bis (RF-64)
Prioridad: Alta
Como voluntario o administrador, quiero contar con una pestaña que liste a todos los participantes del centro, seleccionar una fecha y marcar su asistencia con una observación, para llevar un control diario de concurrencia independiente de los cursos formativos.
Criterios de aceptación:
Dado que accedo a la pestaña de asistencia general, cuando selecciono una fecha, entonces el sistema muestra el listado completo de participantes activos del centro.
Dado el listado de participantes para una fecha, cuando marco su estado de asistencia (Presente / Ausente), entonces el sistema guarda el registro asociado a esa fecha y participante.
El sistema debe permitir agregar una observación de texto libre por cada participante y fecha.
Dado que ya existe un registro de asistencia para un participante en una fecha, cuando vuelvo a esa fecha, entonces el sistema muestra el estado y observación ya cargados, permitiendo su edición.
Este registro es independiente del registro de asistencia por curso; ambos se gestionan y consultan por separado.
El registro de asistencia general debe quedar disponible para consulta en la línea de tiempo del participante.
La carga y modificación de asistencia debe quedar registrada en el log de auditoría.

HU-20 (RF-20)
Prioridad: Alta
Como participante o voluntario, quiero visualizar el progreso formativo calculado según las actividades completadas, para conocer el avance del participante en cada curso.

Criterios de aceptación:
Dado un participante con actividades completadas, cuando se consulta su progreso, entonces el sistema muestra un porcentaje calculado sobre el total de actividades del curso.
El progreso debe actualizarse automáticamente al completarse una nueva actividad.

HU-21 (RF-21)
Prioridad: Media
Como administrador, quiero que el sistema sugiera habilidades adquiridas a partir de los cursos creados mediante IA generativa, para formar el perfil de habilidades de los participantes.

Criterios de aceptación:
Dado un curso completado por un participante, cuando se ejecuta la sugerencia de habilidades, entonces el sistema muestra una lista de habilidades propuestas por Gemini.
Si la API de Gemini falla, el sistema debe integrar la API de OpenAI para sugerir habilidades
El administrador debe poder aceptar o descartar las habilidades sugeridas antes de guardarlas en el perfil.
Si el participante ya tiene alguna habilidad sugerida, la IA no debe ofrecer las mismas.

HU-22 (RF-22)
Prioridad: Media
Como docente externo, quiero consultar los contenidos educativos del centro,
para complementar mi actividad docente.

Criterios de aceptación:
Dado que soy un docente externo, cuando accedo a la sección de contenidos, entonces puedo ver y consultar el material educativo habilitado.
El acceso debe estar limitado a los contenidos autorizados para su rol.

HU-23 (RF-23)
Prioridad: Media
Como docente externo, quiero subir contenido educativo propio, para compartirlo con el centro, sujeto a moderación del administrador.

Criterios de aceptación:
Dado un usuario con rol "Docente Externo", cuando inicia sesión, entonces accede únicamente a la vista de carga/consulta de contenido y no a otras secciones del sistema.
Dado que subo un material, cuando lo envío, entonces queda en estado 'pendiente de moderación'.
El administrador debe poder aprobar o rechazar el contenido antes de que sea visible públicamente.
El docente debe recibir una notificación del resultado de la moderación.
El docente externo solo puede editar o eliminar su propio contenido mientras esté en estado "pendiente".
Una vez aprobado el contenido por el administrador, el docente externo puede consultarlo pero no editarlo.
El contenido aprobado por el administrador queda publicado de forma pública (visible sin necesidad de iniciar sesión, igual que las campañas de concientización).

HU-24 (RF-24)
Prioridad: Alta
Como usuario que consume contenido público de campañas y cursos, quiero dejar comentarios de opinión, para compartir mi feedback sobre el contenido.

Criterios de aceptación:
Dado un contenido público, cuando escribo y envío un comentario, entonces queda publicado asociado a ese contenido.
El sistema debe solicitar Nombre y Apellido de quien comenta
El sistema debe sanitizar el contenido del comentario para prevenir XSS.
El administrador debe poder moderar/eliminar comentarios inapropiados.

HU-25 (RF-25)
Prioridad: Media
Como usuario que descarga contenido de la biblioteca pública, quiero registrar nombre, apellido y un teléfono de contacto, para que el centro pueda hacer seguimiento de mi interés.
Criterios de aceptación:
Dado que solicito descargar un contenido de la biblioteca pública, cuando ingreso mi nombre, apellido y teléfono, entonces se habilita la descarga y el dato queda registrado.
El sistema debe validar el formato del número de teléfono ingresado.


HU-26 (RF-26)
Prioridad: Media
Como participante, quiero recibir notificaciones automáticas por email al inscribirme en un curso y recordatorios previos a la sesión, para no perderme las actividades formativas.



Criterios de aceptación:
Dado que me inscribo en un curso, cuando la inscripción se confirma, entonces recibo un email de confirmación.
El sistema debe enviar un recordatorio automático antes del inicio de cada sesión programada.

HU-27 (RF-27)
Prioridad: Alta
Como usuario, quiero buscar cursos por nombre, para encontrar rápidamente la formación que me interesa.

Criterios de aceptación:
Dado que ingreso un texto en el buscador, cuando el nombre de un curso coincide parcial o totalmente, entonces se muestra en los resultados.
Si no hay coincidencias, el sistema debe mostrar un mensaje indicándolo.

HU-28 (RF-28)
Prioridad: Alta
Como usuario, quiero buscar cursos por área formativa, para filtrar la oferta según mi interés temático.

Criterios de aceptación:
Dado que selecciono un área formativa, cuando aplico el filtro, entonces se listan solo los cursos de esa área.
El sistema debe permitir combinar la búsqueda por nombre y por área.

3. Promoción Humana
HU-29 (RF-29)
Prioridad: Alta
Como administrador o voluntario, quiero crear una ficha de proceso personal por participante, para documentar su situación y acompañamiento de forma integral.

Criterios de aceptación:
Dado un participante sin ficha previa, cuando completo los datos obligatorios, entonces la ficha se crea y queda asociada al participante.
En el caso de que falte algún campo obligatorio de la ficha, informar con mensaje “Completar todos los campos obligatorios para continuar”
Solo los roles administrador y voluntario pueden acceder a esta funcionalidad.

HU-30 (RF-30)
Prioridad: Alta
Como administrador o voluntario, quiero editar la ficha de proceso personal de un participante, para mantener actualizada su información de seguimiento.



Criterios de aceptación:
Dado una ficha existente, cuando modifico sus datos y guardo, entonces los cambios quedan reflejados y aparece mensaje de exito “Ficha personal modificada correctamente”
Cada edición debe quedar registrada en el log de auditoría.

HU-31 (RF-31)
Prioridad: Alta
Como administrador o voluntario, quiero que la ficha de proceso personal tenga campos estructurados (nombre, apellido, edad, habilidades, situación personal, avances),
para registrar información consistente de cada participante.

Criterios de aceptación:
Dado el formulario de ficha, cuando lo completo, entonces todos los campos estructurados definidos están disponibles, son editables y obligatorios
La fecha debe estar en formato Dia/Mes/Año
Los campos obligatorios deben validarse antes de guardar.

HU-32 (RF-32)
Prioridad: Alta
Como administrador o voluntario, quiero registrar hitos de avance con fecha, descripción y responsable, para documentar los logros y eventos relevantes del proceso del participante.

Criterios de aceptación:
Dado una ficha de participante, cuando agrego un hito con fecha, descripción y responsable, entonces queda registrado en su historial.
Los hitos deben poder consultarse ordenados cronológicamente.

HU-33 (RF-33)
Prioridad: Alta
Como voluntario o responsable, quiero visualizar una línea de tiempo cronológica del historial del participante, para tener una visión integral de su evolución.

Criterios de aceptación:
Dado un participante con hitos registrados, cuando accedo a su línea de tiempo, entonces se muestran ordenados cronológicamente con fecha, descripción y responsable.

HU-34 (RF-34)
Prioridad: Media
Como participante,quiero visualizar una línea de tiempo de mi historial de actividad,
para conocer mi propio proceso y avances dentro del centro.

Criterios de aceptación:
Dado que soy participante autenticado, cuando accedo a mi historial, entonces veo únicamente la información autorizada para mi rol.
Los datos confidenciales restringidos a profesionales no deben mostrarse al participante.
HU-35 (RF-35)
Prioridad: Alta
Como administrador, quiero que toda modificación sobre fichas de proceso personal quede registrada en el log de auditoría, para garantizar trazabilidad sobre datos sensibles.

Criterios de aceptación:
Dado que se crea, edita o elimina una ficha, cuando la operación se completa, entonces se registra usuario, fecha y tipo de operación en el log de auditoría.
El log debe ser inmutable y consultable solo por administradores.

HU-36 (RF-36)
Prioridad: Alta
Como voluntario o responsable, quiero buscar participantes por nombre, para acceder rápidamente a su ficha de proceso personal.

Criterios de aceptación:
Dado que ingreso un nombre o parte de él, cuando ejecuto la búsqueda, entonces se listan los participantes cuyo nombre coincide.
Solo se muestran resultados a usuarios con permisos para ver fichas de participantes.

HU-37 (RF-37)
Prioridad: Media
Como voluntario o responsable, quiero exportar en PDF la ficha de descripción y avance de un participante, para compartir o archivar su información de proceso.

Criterios de aceptación:
Dado una ficha de participante, cuando solicito exportarla, entonces el sistema genera un PDF con los datos de la ficha y su línea de avances.
El PDF exportado debe respetar los mismos permisos de acceso que la ficha original.

4. Inserción Laboral

HU-38 (RF-38)
Prioridad: Alta
Como usuario que se registra con rol empresa,quiero que mi cuenta requiera aprobación del administrador, para garantizar que solo empresas validadas publiquen ofertas laborales.

Criterios de aceptación:
Dado que me registro como empresa, cuando envío la solicitud, entonces mi cuenta queda en estado 'pendiente de aprobación'.
Mientras esté pendiente, no debo poder publicar ofertas laborales.
Debo recibir una notificación con el resultado de la aprobación.


HU-39 (RF-39)
Prioridad: Alta
Como empresa,quiero publicar ofertas laborales con título, descripción, ubicación, modalidad, tipo de empleo, salario, horario, requisitos y fecha límite, para dar a conocer oportunidades de empleo a los participantes.

Criterios de aceptación:
Dado que soy una empresa aprobada, cuando completo los datos de la oferta (deben ser campos obligatorios)  y la envío, entonces la oferta se crea en estado pendiente.
El sistema debe validar que todos los campos obligatorios estén completos.

HU-40 (RF-40)
Prioridad: Alta
Como sistema, quiero que las ofertas publicadas queden en estado pendiente hasta su aprobación por un administrador, para asegurar la calidad y pertinencia del contenido visible a los participantes.

Criterios de aceptación:
Dado una oferta recién creada, entonces su estado inicial es 'pendiente' y no es visible públicamente.
Solo tras la aprobación del administrador, la oferta pasa a estado 'publicada' y visible para los participante, voluntario y administradores

HU-41 (RF-41)
Prioridad: Alta
Como administrador, quiero aprobar, rechazar o eliminar ofertas laborales,
para controlar la calidad del contenido publicado en la plataforma.

Criterios de aceptación:
Dado una oferta pendiente, cuando la apruebo, entonces cambia a estado 'publicada' y se hace visible para los participante, voluntario y administradores
Dado una oferta pendiente o publicada, cuando la rechazo o elimino, entonces deja de estar visible y la empresa es notificada.

HU-42 (RF-42)
Prioridad: Alta
Como administrador, quiero eliminar ofertas laborales ya rechazadas o publicadas,
para depurar el catálogo de ofertas obsoletas o inapropiadas.

Criterios de aceptación:
Dado una oferta existente, cuando confirmo su eliminación, entonces deja de estar disponible en el sistema y aparece un cartel de error “Oferta laboral eliminada correctamente”
La eliminación debe quedar registrada en el log de auditoría.


HU-43 (RF-43)
Prioridad: Alta
Como participante,quiero generar automáticamente mi currículum vitae a partir de mi perfil,
para contar con un documento descargable que respalde mi postulación a ofertas.

Criterios de aceptación:
Dado mi perfil completo, cuando solicito generar mi CV, entonces el sistema compila la información y genera un documento descargable.
El sistema debe integrar el servicio de RxResume para la generación de CV.

HU-44 (RF-44)
Prioridad: Media
Como participante, quiero recibir una notificación cuando exista una oferta laboral compatible con mi perfil, para enterarme de oportunidades sin revisar el catálogo constantemente.

Criterios de aceptación:
Dado que se publica una nueva oferta compatible con mi perfil, cuando el sistema procesa la coincidencia, entonces recibo una notificación (email/push).

HU-45 (RF-45)
Prioridad: Alta
Como usuario, quiero buscar ofertas laborales por título, para encontrar rápidamente oportunidades específicas.

Criterios de aceptación:
Dado que ingreso un texto en el buscador, cuando coincide parcial o totalmente con el título de una oferta publicada, entonces se muestra en los resultados.

HU-46 (RF-46)
Prioridad: Alta
Como usuario, quiero buscar ofertas laborales por empresa, para filtrar oportunidades de un empleador específico.

Criterios de aceptación:
Dado que selecciono o escribo el nombre de una empresa, cuando ejecuto la búsqueda, entonces se listan solo las ofertas publicadas por esa empresa.

5. Campañas y Concientización

HU-47 (RF-47)
Prioridad: Alta
Como responsable o voluntario, quiero crear campañas de concientización con imágenes, texto y fecha, para visibilizar las iniciativas ambientales y sociales del centro.

Criterios de aceptación:
Dado que completo los datos de la campaña (imagen, texto, fecha), cuando guardo, entonces la campaña se crea y me aparece mensaje de exito “Campaña creada correctamente”

HU-48 (RF-48)
Prioridad: Alta
Como responsable o voluntario,quiero editar campañas de concientización existentes,
para actualizar su contenido antes o después de su publicación.

Criterios de aceptación:
Dado una campaña existente, cuando modifico sus datos y guardo, entonces los cambios se reflejan correctamente y aparece mensaje de exito “Campaña editada correctamente”

HU-49 (RF-49)
Prioridad: Alta
Como responsable o voluntario, quiero publicar una campaña de concientización,
para hacerla visible al público general.

Criterios de aceptación:
Dado una campaña, cuando la publico, entonces pasa a estar visible en la sección pública del sitio.

HU-50 (RF-50)
Prioridad: Alta
Como visitante no autenticado, quiero ver las campañas publicadas, para conocer las iniciativas del centro sin necesidad de crear una cuenta.

Criterios de aceptación:
Dado que accedo a la sección pública (landing page) sin estar autenticado, cuando ingreso, entonces puedo ver todas las campañas publicadas.
Las campañas no publicadas no deben ser visibles en esta sección.

HU-51 (RF-51)
Prioridad: Alta
Como usuario, quiero buscar campañas por título, para encontrar rápidamente una iniciativa específica.

Criterios de aceptación:
Dado que ingreso un texto en el buscador, cuando coincide parcial o totalmente con el título de una campaña publicada, entonces se muestra en los resultados.

6. Dashboard y Reportes
HU-52 (RF-52)
Prioridad: Alta
Como administrador, quiero ver en el dashboard indicadores clave (participantes activos, cursos en desarrollo, cursos completados, habilidades desarrolladas, inserciones laborales),
para tener una visión general del estado del centro.

Criterios de aceptación:
Dado que accedo al dashboard, cuando se cargan los datos, entonces se muestran los cinco indicadores clave actualizados.

HU-53 (RF-53)
Prioridad: Alta
Como administrador,quiero ver gráficas de evolución temporal de los indicadores principales,para analizar tendencias a lo largo del tiempo.

Criterios de aceptación:
Dado los datos históricos disponibles, cuando accedo al dashboard, entonces se muestran gráficas de evolución por indicador.

HU-54 (RF-54)
Prioridad: Media
Como administrador, quiero filtrar el dashboard por período de tiempo y módulo,
para analizar la información según distintos criterios.

Criterios de aceptación:
Dado que aplico un filtro de fecha, área o módulo, cuando se ejecuta, entonces los indicadores y gráficas se actualizan en consecuencia.

HU-55 (RF-55)
Prioridad: Media
Como administrador, quiero exportar reportes del dashboard en formato PDF, para compartir la información con otros interesados o archivarla.
Criterios de aceptación:
Dado el dashboard con los filtros aplicados, cuando solicito exportar, entonces el sistema genera un PDF con los indicadores y gráficas actuales.

7. Observabilidad
HU-56 (RF-56)
Prioridad: Alta
Como desarrollador del sistema, quiero contar con un módulo de observabilidad, gestionado a traves de una url privada de los desarrolladores, que monitoree el estado y uso de recursos, para detectar fallos, analizar rendimiento y garantizar disponibilidad de la plataforma.

Criterios de aceptación:
Dado que un microservicio genera métricas, cuando se exportan, entonces quedan disponibles en el panel de observabilidad.
El sistema debe permitir visualizar métricas clave (errores, latencia, uso de recursos) por microservicio.
Ante una caída o degradación de un servicio, el sistema debe reflejarlo en el panel de observabilidad.

8. Integraciones Externas
HU-57 (RF-57)
Prioridad: Media
Como usuario, quiero ver la ubicación del centro en un mapa integrado,
para poder llegar fácilmente.

Criterios de aceptación:
Dado que accedo a la página de contacto/ubicación, cuando se carga, entonces se muestra un mapa (Google Maps API) con la ubicación del centro.

HU-58 (RF-58)
Prioridad: Media
Como administrador,quiero que el sistema integre la API generativa de IA (Gemini) para sugerir habilidades, para automatizar y enriquecer el proceso de identificación de competencias adquiridas.

Criterios de aceptación:
Dado un curso o actividad completada, cuando se invoca la integración con Gemini, entonces se reciben sugerencias de habilidades asociadas.
Ante una falla de la API de Gemini, el sistema debe utilizar la API de respaldo configurada (OpenAI).

HU-59 (RF-59)
Prioridad: Alta
Como participante,quiero que el sistema integre un servicio externo de generación de currículums,para obtener un CV profesional a partir de mi información registrada.

Criterios de aceptación:
Dado mi perfil con datos completos, cuando solicito generar mi CV, entonces el sistema se comunica con el servicio de generación de CV y entrega un documento descargable.
Si el servicio no está disponible, el sistema debe informar el error de forma clara al usuario.

9. Sistema de Notificaciones
HU-60 (RF-60)
Prioridad: Alta
Como usuario, quiero recibir notificaciones por correo electrónico ante eventos clave (inscripción a curso, nueva oferta laboral, aprobación de contenido), para mantenerme informado sin tener que revisar la plataforma constantemente.

Criterios de aceptación:
Dado que ocurre un evento clave asociado a mi cuenta, cuando el sistema lo procesa, entonces recibo un email correspondiente.

HU-61 (RF-61)
Prioridad: Media
Como usuario con la aplicación instalada (PWA), quiero recibir notificaciones push,
para enterarme de novedades relevantes en tiempo real.

Criterios de aceptación:
Dado que tengo la PWA instalada y di permiso de notificaciones, cuando ocurre un evento relevante, entonces recibo una notificación push.

HU-62 (RF-62)
Prioridad: Baja
Como usuario, quiero configurar qué tipo de notificaciones deseo recibir, para personalizar mi experiencia y evitar el exceso de comunicaciones.

Criterios de aceptación:
Dado que accedo a la configuración de notificaciones, cuando desactivo un tipo de notificación, entonces dejo de recibir ese tipo de aviso (email y/o push según corresponda).
