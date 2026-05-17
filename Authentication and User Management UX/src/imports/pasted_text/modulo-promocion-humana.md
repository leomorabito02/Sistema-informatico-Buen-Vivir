2.2.3 Módulo de Promoción Humana
Este módulo documenta el proceso personal de cada participante a lo largo del tiempo, desde una perspectiva integral y humanista.
 
¿Qué quiero hacer en este módulo?
→ Registrar una ficha de proceso personal para cada participante, con información cualitativa y cuantitativa.
→ Documentar hitos y avances significativos en el proceso de cada persona.
→ Visualizar una línea de tiempo del historial del participante.
→ Garantizar que estos datos sean estrictamente confidenciales, accesibles solo por administradores y profesionales autorizados.
→ Implementar un motor de búsqueda de personas

 
ID
Descripción del Requerimiento
Prioridad
RF-23
El sistema debe permitir la creación de una ficha de proceso personal por participante, accesible solo para el rol administrador y voluntario
Alta
RF-24
El sistema debe permitir la edición de una ficha de proceso personal por participante, accesible solo para el rol administrador.
Alta
RF-25
La ficha de proceso personal deberá contener campos estructurados para registrar datos generales del participante: nombre, apellido, edad, habilidades, situación personal actual y avances
Alta
RF-26
El sistema debe permitir registrar hitos de avances con fecha, descripción y responsable en el panel del administrador y de voluntarios
Alta
RF-27
El sistema debe mostrar una línea de tiempo cronológica del historial del participante para voluntario y responsable
Alta
RF-28 
El sistema debe mostrar una línea de tiempo cronológica del historial de actividad para el participante


RF-29
Las modificaciones sobre fichas de proceso personal deben quedar registradas en el log de auditoría con usuario, fecha y tipo de operación.
Alta


Implementar un motor de búsqueda de participantes por nombre 
Alta


Implementar un motor de búsqueda de participantes por nombres
Alta


El sistema debe permitir exportar pdf sobre la ficha de descripción y avance de los participantes para los voluntarios y el responsable



2.2.4 Módulo de Inserción Laboral
Este módulo vincula a los participantes con oportunidades laborales publicadas por empresas y voluntarios.
 
¿Qué quiero hacer en este módulo?
→ Permitir a empresas y voluntarios publicar ofertas de trabajo.
→ Moderar las ofertas publicadas antes de que sean visibles para los participantes.
→ Vincular participantes con ofertas según su perfil de habilidades.
→ Integrar una servicio interno para generar el currículum vitae del participante automáticamente.
→ Implementar un motor de búsqueda de ofertas laborales

 
ID
Descripción del Requerimiento
Prioridad
RF-30
El sistema debe permitir al  role empresa publicar ofertas laborales con título, descripción, requisitos y fecha límite.
Alta
RF-31
Las ofertas publicadas deben quedar en estado pendiente hasta su aprobación por el administrador.
Alta
RF-32
El administrador debe poder aprobar, rechazar o eliminar ofertas laborales.
Alta
RF-33
El administrador debe poder rechazar eliminar ofertas laborales.
Alta
RF-34
El sistema debe sugerir ofertas a participantes según su perfil de habilidades y formación.
Media
RF-35
El sistema debe integrar un servicio interno de generación de CV que compile el perfil del participante en un documento descargable.
Alta
RF-36
El participante debe recibir una notificación cuando exista una oferta compatible con su perfil.
Media


Implementar un motor de búsqueda de ofertas laborales por título 
Alta


Implementar un motor de búsqueda de ofertas laborales por empresa
Alta

 

2.2.5 Módulo de Campañas y Concientización
Este módulo permite al centro visibilizar sus iniciativas de concientización ambiental y social.
 
ID
Descripción del Requerimiento
Prioridad
RF-37
El sistema debe permitir, al rol de responsables y voluntarios, crear campañas de concientización con imágenes, texto y fecha.
Alta
RF-39
El sistema debe permitir, al rol de responsables y voluntarios, editar campañas de concientización con imágenes, texto y fecha.
Alta
RF-40
El sistema debe permitir, al rol de responsables y voluntarios, publicar campañas de concientización con imágenes, texto y fecha.
Alta
RF-41
Las campañas publicadas deben ser visibles para usuarios no autenticados (sección pública del sitio).
Alta


Implementar un motor de búsqueda de campañas según título 
Alta

 
2.2.6 Dashboard y Reportes
El dashboard provee al equipo directivo del centro una visión analítica del estado del sistema.
 
¿Qué quiero hacer en este módulo?
→ Visualizar la evolución y progreso formativo de los participantes.
→ Monitorear el desarrollo de habilidades y la inserción laboral.
→ Exportar reportes para la toma de decisiones

 
ID
Descripción del Requerimiento
Prioridad
RF-42
El dashboard debe mostrar indicadores clave: cantidad de participantes activos, cursos en desarrollo, habilidades desarrolladas e inserciones laborales.
Alta
RF-43
El dashboard debe mostrar gráficas de evolución temporal de los indicadores principales.
Alta
RF-44
El sistema debe permitir filtrar el dashboard por período de tiempo, área formativa y módulo.
Media
RF-45
El sistema debe permitir exportar reportes en formato PDF.
Media


2.2.7 Módulo de observabilidad
ID
Descripción del Requerimiento
Prioridad
RF-51
Como administrador del sistema, quiero contar con un módulo de observabilidad que me permita monitorear el estado y el uso de recursos del sistema, para detectar fallos, analizar el rendimiento y garantizar la disponibilidad de la plataforma. 
Alta


2.2.8 Integraciones Externas
ID
Descripción del Requerimiento
Prioridad
RF-47
El sistema debe integrar una API de mapas para mostrar la ubicación del centro
Media
RF-48
El sistema debe integrar una pasarela de pagos (ej. Mercado Pago) para recibir donaciones de forma segura.
Alta
RF-49
El sistema debe integrar la API generativa de IA (ej. Anthropic Claude API o similar) para la sugerencia de habilidades.
Media
RF-50
El sistema debe integrar una API externa de generación de currículums vitae.
Alta

 
2.2.9 Sistema de Notificaciones
ID
Descripción del Requerimiento
Prioridad
RF-51
El sistema debe enviar notificaciones por correo electrónico ante eventos clave: inscripción a curso, nueva oferta laboral, aprobación de contenido.
Alta
RF-52
El sistema debe implementar notificaciones push (PWA) para usuarios con la aplicación instalada.
Media
RF-53
Los usuarios deben poder configurar qué tipo de notificaciones desea recibir.
Baja

