import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { Leaf, ArrowLeft, Clock, Users, BookOpen, CheckCircle, Circle, LogOut, User, Calendar, FileText, Download } from 'lucide-react';

interface CourseDetailProps {
  userRole: 'admin' | 'participante' | 'voluntario' | 'empresa' | 'docente';
  onLogout: () => void;
}

export function CourseDetail({ userRole, onLogout }: CourseDetailProps) {
  const { id } = useParams();
  const [isEnrolled, setIsEnrolled] = useState(false);
  const [showEnrollModal, setShowEnrollModal] = useState(false);

  const course = {
    id: 1,
    title: 'Panadería Artesanal',
    description: 'Aprende las técnicas tradicionales y modernas de panadería. Este curso te enseñará desde los fundamentos básicos hasta técnicas avanzadas de fermentación y horneado.',
    area: 'Gastronomía',
    teacher: 'María Rodríguez',
    teacherBio: 'Panadera con 15 años de experiencia en panadería artesanal',
    duration: '8 semanas',
    enrolled: 24,
    maxStudents: 30,
    image: '🍞',
    startDate: '15 de Junio, 2026',
    schedule: 'Martes y Jueves, 14:00 - 17:00',
    modules: [
      {
        id: 1,
        title: 'Introducción a la Panadería',
        description: 'Conceptos básicos, ingredientes y herramientas',
        activities: [
          { id: 1, title: 'Video: Historia de la panadería', completed: true },
          { id: 2, title: 'Lectura: Tipos de harinas', completed: true },
          { id: 3, title: 'Quiz: Ingredientes básicos', completed: false },
        ]
      },
      {
        id: 2,
        title: 'Masas Básicas',
        description: 'Preparación de masas madre y masas directas',
        activities: [
          { id: 4, title: 'Video: Preparación de masa madre', completed: false },
          { id: 5, title: 'Práctica: Primer amasado', completed: false },
          { id: 6, title: 'Tarea: Documentar proceso', completed: false },
        ]
      },
      {
        id: 3,
        title: 'Técnicas de Fermentación',
        description: 'Control de temperatura y tiempos de fermentación',
        activities: [
          { id: 7, title: 'Video: Fermentación controlada', completed: false },
          { id: 8, title: 'Práctica: Fermentación larga', completed: false },
        ]
      },
      {
        id: 4,
        title: 'Horneado',
        description: 'Técnicas de horneado y acabados',
        activities: [
          { id: 9, title: 'Video: Control de temperatura', completed: false },
          { id: 10, title: 'Práctica: Horneado de pan básico', completed: false },
        ]
      },
    ],
    materials: [
      { name: 'Guía de Ingredientes.pdf', size: '2.4 MB' },
      { name: 'Recetario Básico.pdf', size: '5.1 MB' },
      { name: 'Video Tutorial - Masa Madre.mp4', size: '124 MB' },
    ]
  };

  const totalActivities = course.modules.reduce((sum, module) => sum + module.activities.length, 0);
  const completedActivities = course.modules.reduce(
    (sum, module) => sum + module.activities.filter(a => a.completed).length,
    0
  );
  const progress = Math.round((completedActivities / totalActivities) * 100);

  const handleEnroll = () => {
    if (userRole === 'participante') {
      setShowEnrollModal(true);
    } else {
      setIsEnrolled(true);
    }
  };

  const confirmEnroll = () => {
    setIsEnrolled(true);
    setShowEnrollModal(false);
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="bg-card border-b border-border p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link to="/dashboard" className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center">
              <Leaf className="w-6 h-6 text-primary-foreground" />
            </Link>
            <div>
              <h1 className="text-2xl">Detalle del Curso</h1>
              <p className="text-sm text-muted-foreground">Información completa y módulos</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Link
              to="/courses"
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
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Main Column */}
          <div className="lg:col-span-2 space-y-6">
            {/* Course Header */}
            <div className="bg-card rounded-lg border border-border overflow-hidden">
              <div className="aspect-video bg-gradient-to-br from-primary to-[#1a3a1f] flex items-center justify-center">
                <span className="text-9xl">{course.image}</span>
              </div>
              <div className="p-6">
                <div className="flex items-center gap-3 mb-4">
                  <span className="inline-flex px-3 py-1 rounded-full text-sm bg-primary/10 text-primary">
                    {course.area}
                  </span>
                  {isEnrolled && (
                    <span className="inline-flex px-3 py-1 rounded-full text-sm bg-[#66bb6a]/10 text-[#2d7a3e]">
                      Inscrito
                    </span>
                  )}
                </div>
                <h2 className="text-3xl mb-3">{course.title}</h2>
                <p className="text-muted-foreground mb-6">{course.description}</p>

                <div className="grid grid-cols-2 gap-4 mb-6">
                  <div className="flex items-center gap-3">
                    <User className="w-5 h-5 text-primary" />
                    <div>
                      <p className="text-sm text-muted-foreground">Docente</p>
                      <p>{course.teacher}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Clock className="w-5 h-5 text-primary" />
                    <div>
                      <p className="text-sm text-muted-foreground">Duración</p>
                      <p>{course.duration}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Calendar className="w-5 h-5 text-primary" />
                    <div>
                      <p className="text-sm text-muted-foreground">Inicio</p>
                      <p>{course.startDate}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Users className="w-5 h-5 text-primary" />
                    <div>
                      <p className="text-sm text-muted-foreground">Cupos</p>
                      <p>{course.enrolled}/{course.maxStudents}</p>
                    </div>
                  </div>
                </div>

                {!isEnrolled && (userRole === 'participante' || userRole === 'voluntario') && (
                  <button
                    onClick={handleEnroll}
                    className="w-full py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
                  >
                    Inscribirse al Curso
                  </button>
                )}
              </div>
            </div>

            {/* Course Modules */}
            {isEnrolled && (
              <div className="bg-card rounded-lg border border-border p-6">
                <div className="flex items-center justify-between mb-6">
                  <h3 className="text-2xl">Módulos del Curso</h3>
                  <div className="text-right">
                    <p className="text-sm text-muted-foreground">Progreso total</p>
                    <p className="text-2xl text-primary">{progress}%</p>
                  </div>
                </div>

                <div className="w-full bg-secondary rounded-full h-3 mb-8 overflow-hidden">
                  <div
                    className="h-full bg-primary transition-all"
                    style={{ width: `${progress}%` }}
                  />
                </div>

                <div className="space-y-6">
                  {course.modules.map((module, index) => (
                    <div key={module.id} className="border border-border rounded-lg overflow-hidden">
                      <div className="bg-muted p-4">
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center">
                            {index + 1}
                          </div>
                          <div className="flex-1">
                            <h4 className="mb-1">{module.title}</h4>
                            <p className="text-sm text-muted-foreground">{module.description}</p>
                          </div>
                          <div className="text-sm text-muted-foreground">
                            {module.activities.filter(a => a.completed).length}/{module.activities.length}
                          </div>
                        </div>
                      </div>
                      <div className="p-4 space-y-2">
                        {module.activities.map((activity) => (
                          <div
                            key={activity.id}
                            className="flex items-center gap-3 p-3 rounded-lg hover:bg-muted transition-colors"
                          >
                            {activity.completed ? (
                              <CheckCircle className="w-5 h-5 text-primary flex-shrink-0" />
                            ) : (
                              <Circle className="w-5 h-5 text-muted-foreground flex-shrink-0" />
                            )}
                            <span className={activity.completed ? 'text-muted-foreground line-through' : ''}>
                              {activity.title}
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Teacher Info */}
            <div className="bg-card rounded-lg border border-border p-6">
              <h3 className="text-xl mb-4">Sobre el Docente</h3>
              <div className="flex items-center gap-3 mb-3">
                <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center">
                  <User className="w-6 h-6 text-primary" />
                </div>
                <div>
                  <p>{course.teacher}</p>
                  <p className="text-sm text-muted-foreground">{course.teacherBio}</p>
                </div>
              </div>
            </div>

            {/* Schedule */}
            <div className="bg-card rounded-lg border border-border p-6">
              <h3 className="text-xl mb-4">Horario</h3>
              <div className="flex items-start gap-3">
                <Calendar className="w-5 h-5 text-primary flex-shrink-0 mt-1" />
                <div>
                  <p className="mb-1">{course.schedule}</p>
                  <p className="text-sm text-muted-foreground">Inicia: {course.startDate}</p>
                </div>
              </div>
            </div>

            {/* Materials */}
            <div className="bg-card rounded-lg border border-border p-6">
              <h3 className="text-xl mb-4">Materiales del Curso</h3>
              <div className="space-y-3">
                {course.materials.map((material, index) => (
                  <div
                    key={index}
                    className="flex items-center justify-between p-3 rounded-lg border border-border hover:border-primary transition-colors"
                  >
                    <div className="flex items-center gap-3 flex-1 min-w-0">
                      <FileText className="w-5 h-5 text-primary flex-shrink-0" />
                      <div className="min-w-0">
                        <p className="text-sm truncate">{material.name}</p>
                        <p className="text-xs text-muted-foreground">{material.size}</p>
                      </div>
                    </div>
                    <button className="p-2 rounded-lg hover:bg-secondary transition-colors flex-shrink-0">
                      <Download className="w-4 h-4" />
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* Enrollment Confirmation Modal */}
      {showEnrollModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-md p-6">
            <h2 className="text-2xl mb-4">Confirmar Inscripción</h2>
            <p className="text-muted-foreground mb-6">
              ¿Estás seguro que deseas inscribirte al curso "{course.title}"?
            </p>
            <div className="bg-secondary rounded-lg p-4 mb-6">
              <p className="text-sm mb-2">
                <strong className="text-foreground">Horario:</strong> {course.schedule}
              </p>
              <p className="text-sm">
                <strong className="text-foreground">Inicio:</strong> {course.startDate}
              </p>
            </div>
            <div className="flex gap-3">
              <button
                onClick={() => setShowEnrollModal(false)}
                className="flex-1 px-6 py-3 rounded-lg border border-border hover:bg-secondary transition-colors"
              >
                Cancelar
              </button>
              <button
                onClick={confirmEnroll}
                className="flex-1 px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
              >
                Confirmar
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
