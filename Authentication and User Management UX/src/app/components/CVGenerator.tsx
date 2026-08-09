import { useState } from 'react';
import { Link } from 'react-router-dom';
import { LogOut, Download, FileText, User, Mail, Phone, MapPin, GraduationCap, Award, AlertCircle, ArrowLeft } from 'lucide-react';
import { Toast, ToastState } from './base';

interface CVGeneratorProps {
  onLogout: () => void;
}

export function CVGenerator({ onLogout }: CVGeneratorProps) {
  const [showPreview, setShowPreview] = useState(false);
  const [toast, setToast] = useState<ToastState | null>(null);
  const [serviceError, setServiceError] = useState(false);

  const participantData = {
    personalInfo: {
      name: 'María González',
      email: 'maria.gonzalez@email.com',
      phone: '+54 9 3547 123-456',
      location: 'Alta Gracia, Córdoba, Argentina',
      age: 28,
      photo: null
    },
    summary: 'Persona responsable y comprometida con amplia experiencia en panadería artesanal. Busco oportunidades para aplicar mis habilidades técnicas y crecer profesionalmente en un ambiente de trabajo colaborativo.',
    skills: [
      'Panadería Artesanal',
      'Trabajo en Equipo',
      'Responsabilidad',
      'Puntualidad',
      'Computación Básica',
      'Organización',
      'Atención al Cliente'
    ],
    courses: [
      {
        title: 'Panadería Artesanal',
        institution: 'Centro Buen Vivir',
        date: 'Marzo 2026 - Mayo 2026',
        status: 'Completado',
        grade: 'Excelente'
      },
      {
        title: 'Computación Básica',
        institution: 'Centro Buen Vivir',
        date: 'Abril 2026 - Presente',
        status: 'En curso',
        grade: 'En progreso'
      },
      {
        title: 'Costura y Diseño',
        institution: 'Centro Buen Vivir',
        date: 'Enero 2026 - Marzo 2026',
        status: 'Completado',
        grade: 'Muy Bueno'
      }
    ],
    experience: [
      {
        position: 'Ayudante de Panadería',
        company: 'Panadería Familiar',
        period: '2024 - 2025',
        description: 'Asistencia en preparación de masas, horneado y atención al público'
      }
    ],
    languages: [
      { name: 'Español', level: 'Nativo' }
    ]
  };

  const handleGenerateCV = () => {
    setServiceError(false);
    setShowPreview(true);
  };

  const handleDownloadPDF = () => {
    // Mock: simula una falla ocasional del servicio externo RxResume para ilustrar el manejo de error (HU-59).
    const rxResumeAvailable = Math.random() > 0.15;
    if (!rxResumeAvailable) {
      setServiceError(true);
      setToast({ type: 'error', message: 'El servicio de generación de CV (RxResume) no está disponible en este momento. Intentá nuevamente en unos minutos.' });
      return;
    }
    setServiceError(false);
    setToast({ type: 'success', message: 'Tu CV fue generado con RxResume y está listo para descargar.' });
  };

  return (
    <div className="min-h-screen bg-background">
      <Toast toast={toast} onClose={() => setToast(null)} />

      {/* Header */}
      <header className="bg-card border-b border-border p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link to="/dashboard" className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center">
              <img src="/logo.png" alt="Buen Vivir" className="w-8 h-8 object-contain" />
            </Link>
            <div>
              <h1 className="text-2xl">Generador de Currículum Vitae</h1>
              <p className="text-sm text-muted-foreground">Crea tu CV profesional automáticamente</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Link
              to="/dashboard"
              className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors flex items-center gap-2"
            >
              <ArrowLeft className="w-4 h-4" />
              Volver
            </Link>
            <button
              onClick={onLogout}
              className="px-4 py-2 rounded-lg bg-destructive text-destructive-foreground hover:opacity-90 transition-opacity flex items-center gap-2"
            >
              <LogOut className="w-4 h-4" />
              Salir
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="p-6">
        <div className="max-w-5xl mx-auto">
          {!showPreview ? (
            <>
              {/* Information Notice */}
              <div className="bg-primary/10 border border-primary/20 rounded-lg p-6 mb-8">
                <div className="flex items-start gap-4">
                  <FileText className="w-6 h-6 text-primary flex-shrink-0 mt-1" />
                  <div>
                    <h3 className="text-lg mb-2">¿Cómo funciona el generador de CV?</h3>
                    <p className="text-muted-foreground mb-2">
                      El sistema se integra con <strong className="text-foreground">RxResume</strong>, un servicio externo de generación de currículums,
                      y compila automáticamente tu CV a partir de:
                    </p>
                    <ul className="list-disc list-inside space-y-1 text-muted-foreground">
                      <li>Tu información personal registrada en el sistema</li>
                      <li>Los cursos que has completado y estás cursando</li>
                      <li>Las habilidades que has desarrollado</li>
                      <li>Tu experiencia laboral previa (si la has registrado)</li>
                    </ul>
                  </div>
                </div>
              </div>

              {/* Profile Summary */}
              <div className="bg-card rounded-lg border border-border p-8 mb-8">
                <h2 className="text-2xl mb-6">Datos del Perfil</h2>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
                  <div className="flex items-center gap-3">
                    <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                      <User className="w-6 h-6 text-primary" />
                    </div>
                    <div>
                      <p className="text-sm text-muted-foreground">Nombre completo</p>
                      <p>{participantData.personalInfo.name}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                      <Mail className="w-6 h-6 text-primary" />
                    </div>
                    <div>
                      <p className="text-sm text-muted-foreground">Email</p>
                      <p>{participantData.personalInfo.email}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                      <Phone className="w-6 h-6 text-primary" />
                    </div>
                    <div>
                      <p className="text-sm text-muted-foreground">Teléfono</p>
                      <p>{participantData.personalInfo.phone}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                      <MapPin className="w-6 h-6 text-primary" />
                    </div>
                    <div>
                      <p className="text-sm text-muted-foreground">Ubicación</p>
                      <p>{participantData.personalInfo.location}</p>
                    </div>
                  </div>
                </div>

                <div className="mb-6">
                  <h3 className="mb-3 flex items-center gap-2">
                    <GraduationCap className="w-5 h-5 text-primary" />
                    Cursos Realizados ({participantData.courses.length})
                  </h3>
                  <div className="space-y-2">
                    {participantData.courses.map((course, index) => (
                      <div key={index} className="p-3 rounded-lg bg-muted">
                        <div className="flex items-start justify-between">
                          <div>
                            <p className="mb-1">{course.title}</p>
                            <p className="text-sm text-muted-foreground">{course.institution} • {course.date}</p>
                          </div>
                          <span className={`inline-flex px-3 py-1 rounded-full text-xs ${
                            course.status === 'Completado'
                              ? 'bg-primary/10 text-primary'
                              : 'bg-secondary/10 text-secondary'
                          }`}>
                            {course.status}
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                <div>
                  <h3 className="mb-3 flex items-center gap-2">
                    <Award className="w-5 h-5 text-primary" />
                    Habilidades ({participantData.skills.length})
                  </h3>
                  <div className="flex flex-wrap gap-2">
                    {participantData.skills.map((skill, index) => (
                      <span
                        key={index}
                        className="inline-flex px-3 py-1 rounded-full text-sm bg-primary/10 text-primary"
                      >
                        {skill}
                      </span>
                    ))}
                  </div>
                </div>
              </div>

              {/* Generate Button */}
              <div className="text-center">
                <button
                  onClick={handleGenerateCV}
                  className="px-12 py-4 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity text-lg flex items-center gap-3 mx-auto"
                >
                  <FileText className="w-6 h-6" />
                  Generar Currículum Vitae
                </button>
              </div>
            </>
          ) : (
            <>
              {/* CV Preview */}
              <div className="bg-card rounded-lg border border-border mb-6">
                <div className="border-b border-border p-6 flex items-center justify-between">
                  <h2 className="text-2xl">Vista Previa del CV</h2>
                  <button
                    onClick={handleDownloadPDF}
                    className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity flex items-center gap-2"
                  >
                    <Download className="w-5 h-5" />
                    Descargar PDF
                  </button>
                </div>

                {serviceError && (
                  <div className="flex items-start gap-2 text-sm text-destructive bg-destructive/10 border-b border-destructive/20 p-4">
                    <AlertCircle className="w-4 h-4 flex-shrink-0 mt-0.5" />
                    <span>No pudimos generar el PDF porque el servicio RxResume no respondió. Probá de nuevo en unos minutos.</span>
                  </div>
                )}

                <div className="p-12 bg-white text-black">
                  {/* CV Content */}
                  <div className="max-w-3xl mx-auto">
                    {/* Header */}
                    <div className="text-center mb-8 pb-6 border-b-2 border-gray-300">
                      <h1 className="text-4xl mb-2" style={{ color: '#485839' }}>
                        {participantData.personalInfo.name}
                      </h1>
                      <div className="flex flex-wrap justify-center gap-4 text-sm text-gray-600">
                        <span>{participantData.personalInfo.email}</span>
                        <span>•</span>
                        <span>{participantData.personalInfo.phone}</span>
                        <span>•</span>
                        <span>{participantData.personalInfo.location}</span>
                      </div>
                    </div>

                    {/* Profile Summary */}
                    <div className="mb-6">
                      <h2 className="text-xl mb-3" style={{ color: '#485839' }}>PERFIL PROFESIONAL</h2>
                      <p className="text-gray-700">{participantData.summary}</p>
                    </div>

                    {/* Skills */}
                    <div className="mb-6">
                      <h2 className="text-xl mb-3" style={{ color: '#485839' }}>HABILIDADES</h2>
                      <div className="grid grid-cols-2 gap-2">
                        {participantData.skills.map((skill, index) => (
                          <div key={index} className="flex items-center gap-2">
                            <span className="w-2 h-2 rounded-full" style={{ backgroundColor: '#485839' }} />
                            <span className="text-gray-700">{skill}</span>
                          </div>
                        ))}
                      </div>
                    </div>

                    {/* Education */}
                    <div className="mb-6">
                      <h2 className="text-xl mb-3" style={{ color: '#485839' }}>FORMACIÓN</h2>
                      <div className="space-y-4">
                        {participantData.courses.map((course, index) => (
                          <div key={index}>
                            <h3 className="font-medium text-gray-900">{course.title}</h3>
                            <p className="text-gray-600">{course.institution} | {course.date}</p>
                            <p className="text-gray-600">Calificación: {course.grade}</p>
                          </div>
                        ))}
                      </div>
                    </div>

                    {/* Experience */}
                    {participantData.experience.length > 0 && (
                      <div className="mb-6">
                        <h2 className="text-xl mb-3" style={{ color: '#485839' }}>EXPERIENCIA LABORAL</h2>
                        <div className="space-y-4">
                          {participantData.experience.map((exp, index) => (
                            <div key={index}>
                              <h3 className="font-medium text-gray-900">{exp.position}</h3>
                              <p className="text-gray-600">{exp.company} | {exp.period}</p>
                              <p className="text-gray-700 mt-1">{exp.description}</p>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Languages */}
                    <div>
                      <h2 className="text-xl mb-3" style={{ color: '#485839' }}>IDIOMAS</h2>
                      <div className="space-y-1">
                        {participantData.languages.map((lang, index) => (
                          <p key={index} className="text-gray-700">
                            {lang.name}: {lang.level}
                          </p>
                        ))}
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div className="text-center">
                <button
                  onClick={() => setShowPreview(false)}
                  className="px-6 py-3 rounded-lg border border-border hover:bg-secondary transition-colors"
                >
                  Volver a Editar
                </button>
              </div>
            </>
          )}
        </div>
      </main>
    </div>
  );
}
