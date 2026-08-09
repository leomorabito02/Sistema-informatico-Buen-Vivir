import { Link } from 'react-router-dom';
import { useState } from 'react';
import { LogOut, BookOpen, Award, TrendingUp, Sparkles, Download, Calendar, X, ChefHat, Laptop, Sprout, CheckCircle2, Star, Plus, ArrowLeft } from 'lucide-react';
import { Button } from 'primereact/button';

interface ProgressTrackingProps {
  userRole: 'admin' | 'participante' | 'voluntario' | 'empresa' | 'docente';
  onLogout: () => void;
}

export function ProgressTracking({ userRole, onLogout }: ProgressTrackingProps) {
  const [showTimeline, setShowTimeline] = useState(false);

  const enrolledCourses = [
    {
      id: 1,
      title: 'Panadería Artesanal',
      area: 'Gastronomía',
      progress: 75,
      totalActivities: 10,
      completedActivities: 7,
      lastActivity: 'Hace 2 horas',
      nextActivity: 'Técnicas de horneado',
      icon: ChefHat
    },
    {
      id: 2,
      title: 'Computación Básica',
      area: 'Tecnología',
      progress: 60,
      totalActivities: 15,
      completedActivities: 9,
      lastActivity: 'Hace 1 día',
      nextActivity: 'Hojas de cálculo avanzadas',
      icon: Laptop
    },
    {
      id: 3,
      title: 'Huerta Orgánica',
      area: 'Agricultura',
      progress: 40,
      totalActivities: 8,
      completedActivities: 3,
      lastActivity: 'Hace 3 días',
      nextActivity: 'Compostaje y fertilización',
      icon: Sprout
    },
  ];

  const certificates = [
    {
      id: 1,
      courseName: 'Costura y Diseño',
      completedDate: '15 de Abril, 2026',
      grade: 'Excelente'
    },
    {
      id: 2,
      courseName: 'Cocina Saludable',
      completedDate: '22 de Marzo, 2026',
      grade: 'Muy Bueno'
    },
  ];

  const suggestedSkills = [
    'Panadería Artesanal',
    'Gestión del Tiempo',
    'Trabajo en Equipo',
    'Habilidades Digitales',
    'Agricultura Sostenible',
    'Nutrición',
    'Diseño de Moda',
    'Cocina',
  ];

  const timeline = [
    {
      id: 1,
      date: '15 de Mayo, 2026',
      event: 'Completó el curso "Panadería Artesanal"',
      type: 'completion'
    },
    {
      id: 2,
      date: '10 de Mayo, 2026',
      event: 'Alcanzó 75% de progreso en "Panadería Artesanal"',
      type: 'milestone'
    },
    {
      id: 3,
      date: '5 de Mayo, 2026',
      event: 'Se inscribió en "Computación Básica"',
      type: 'enrollment'
    },
    {
      id: 4,
      date: '1 de Mayo, 2026',
      event: 'Completó el curso "Cocina Saludable"',
      type: 'completion'
    },
    {
      id: 5,
      date: '25 de Abril, 2026',
      event: 'Se inscribió en "Huerta Orgánica"',
      type: 'enrollment'
    },
  ];

  const totalProgress = Math.round(
    enrolledCourses.reduce((sum, course) => sum + course.progress, 0) / enrolledCourses.length
  );

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="bg-card border-b border-border p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link to="/dashboard" className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center">
              <img src="/logo.png" alt="Buen Vivir" className="w-8 h-8 object-contain" />
            </Link>
            <div>
              <h1 className="text-2xl">Mi Progreso</h1>
              <p className="text-sm text-muted-foreground">Seguimiento de tu aprendizaje y logros</p>
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
        {/* Stats Overview */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div className="bg-card rounded-lg border border-border p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-lg bg-primary flex items-center justify-center">
                <BookOpen className="w-6 h-6 text-primary-foreground" />
              </div>
              <span className="text-3xl">{enrolledCourses.length}</span>
            </div>
            <h3 className="text-sm text-muted-foreground">Cursos Activos</h3>
          </div>

          <div className="bg-card rounded-lg border border-border p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-lg bg-secondary flex items-center justify-center">
                <TrendingUp className="w-6 h-6 text-white" />
              </div>
              <span className="text-3xl">{totalProgress}%</span>
            </div>
            <h3 className="text-sm text-muted-foreground">Progreso Promedio</h3>
          </div>

          <div className="bg-card rounded-lg border border-border p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-lg bg-[#1D1D1B] flex items-center justify-center">
                <Award className="w-6 h-6 text-white" />
              </div>
              <span className="text-3xl">{certificates.length}</span>
            </div>
            <h3 className="text-sm text-muted-foreground">Certificados</h3>
          </div>

          <div className="bg-card rounded-lg border border-border p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-lg bg-primary/70 flex items-center justify-center">
                <Sparkles className="w-6 h-6 text-white" />
              </div>
              <span className="text-3xl">{suggestedSkills.length}</span>
            </div>
            <h3 className="text-sm text-muted-foreground">Habilidades</h3>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Enrolled Courses Progress */}
          <div className="lg:col-span-2 space-y-6">
            <div className="bg-card rounded-lg border border-border p-6">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-2xl">Cursos en Progreso</h2>
                <button
                  onClick={() => setShowTimeline(true)}
                  className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors flex items-center gap-2"
                >
                  <Calendar className="w-4 h-4" />
                  Ver Timeline
                </button>
              </div>
              <div className="space-y-4">
                {enrolledCourses.map((course) => (
                  <Link
                    key={course.id}
                    to={`/courses/${course.id}`}
                    className="block p-4 rounded-lg border border-border hover:border-primary hover:bg-primary/5 transition-all"
                  >
                    <div className="flex items-start gap-4">
                      <div className="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center flex-shrink-0">
                        <course.icon className="w-6 h-6 text-primary" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-start justify-between mb-2">
                          <div className="flex-1 min-w-0">
                            <h3 className="mb-1">{course.title}</h3>
                            <p className="text-sm text-muted-foreground">{course.area}</p>
                          </div>
                          <span className="text-2xl text-primary ml-3">{course.progress}%</span>
                        </div>

                        <div className="w-full bg-secondary rounded-full h-2 mb-3 overflow-hidden">
                          <div
                            className="h-full bg-primary transition-all"
                            style={{ width: `${course.progress}%` }}
                          />
                        </div>

                        <div className="flex items-center justify-between text-sm">
                          <span className="text-muted-foreground">
                            {course.completedActivities}/{course.totalActivities} actividades completadas
                          </span>
                          <span className="text-muted-foreground">
                            Última actividad: {course.lastActivity}
                          </span>
                        </div>

                        <div className="mt-3 pt-3 border-t border-border">
                          <p className="text-sm">
                            <span className="text-muted-foreground">Próxima actividad:</span>{' '}
                            <span className="text-primary">{course.nextActivity}</span>
                          </p>
                        </div>
                      </div>
                    </div>
                  </Link>
                ))}
              </div>
            </div>

            {/* Certificates */}
            <div className="bg-card rounded-lg border border-border p-6">
              <h2 className="text-2xl mb-6">Mis Certificados</h2>
              <div className="space-y-4">
                {certificates.map((cert) => (
                  <div
                    key={cert.id}
                    className="p-4 rounded-lg border border-border bg-gradient-to-r from-primary/5 to-transparent"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center">
                          <Award className="w-6 h-6 text-primary" />
                        </div>
                        <div>
                          <h3 className="mb-1">{cert.courseName}</h3>
                          <p className="text-sm text-muted-foreground">
                            Completado: {cert.completedDate}
                          </p>
                          <p className="text-sm text-primary">
                            Calificación: {cert.grade}
                          </p>
                        </div>
                      </div>
                      <Button icon={<Download className="w-4 h-4 mr-2" />} label="Exportar" className="p-button-outlined p-button-secondary bg-transparent text-foreground border-border hover:bg-secondary" />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* AI Suggested Skills */}
            <div className="bg-card rounded-lg border border-border p-6">
              <div className="flex items-center gap-2 mb-4">
                <Sparkles className="w-5 h-5 text-primary" />
                <h2 className="text-xl">Habilidades Adquiridas</h2>
              </div>
              <p className="text-sm text-muted-foreground mb-4">
                Sugerencias generadas por IA basadas en tus cursos completados
              </p>
              <div className="flex flex-wrap gap-2">
                {suggestedSkills.map((skill, index) => (
                  <span
                    key={index}
                    className="inline-flex px-3 py-1 rounded-full text-sm bg-primary/10 text-primary border border-primary/20"
                  >
                    {skill}
                  </span>
                ))}
              </div>
            </div>
            
          </div>
        </div>
      </main>

      {/* Timeline Modal */}
      {showTimeline && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-card border-b border-border p-6 flex items-center justify-between">
              <div>
                <h2 className="text-2xl mb-1">Cronología de Avances</h2>
                <p className="text-sm text-muted-foreground">Tu historial de logros y eventos</p>
              </div>
              <button
                onClick={() => setShowTimeline(false)}
                className="p-2 rounded-lg hover:bg-secondary transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6">
              <div className="space-y-6">
                {timeline.map((item, index) => (
                  <div key={item.id} className="flex gap-4">
                    <div className="flex flex-col items-center">
                      <div className={`w-10 h-10 rounded-full flex items-center justify-center text-primary-foreground flex-shrink-0 ${
                        item.type === 'completion' ? 'bg-primary' :
                        item.type === 'milestone' ? 'bg-secondary' :
                        'bg-[#1D1D1B]'
                      }`}>
                        {item.type === 'completion' && <CheckCircle2 className="w-5 h-5" />}
                        {item.type === 'milestone' && <Star className="w-5 h-5" />}
                        {item.type === 'enrollment' && <Plus className="w-5 h-5" />}
                      </div>
                      {index < timeline.length - 1 && (
                        <div className="w-0.5 flex-1 bg-border mt-2" />
                      )}
                    </div>
                    <div className="flex-1 pb-6">
                      <div className="flex items-start justify-between mb-2">
                        <h3 className="text-base font-medium">{item.event}</h3>
                        <span className="text-sm text-muted-foreground whitespace-nowrap ml-4">{item.date}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
