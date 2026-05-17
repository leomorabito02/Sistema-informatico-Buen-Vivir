import { Link } from 'react-router-dom';
import { Leaf, LogOut, BookOpen, Award, TrendingUp, Sparkles, Download } from 'lucide-react';

interface ProgressTrackingProps {
  userRole: 'admin' | 'participante' | 'voluntario' | 'empresa' | 'docente';
  onLogout: () => void;
}

export function ProgressTracking({ userRole, onLogout }: ProgressTrackingProps) {
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
      image: '🍞'
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
      image: '💻'
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
      image: '🌱'
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
              <Leaf className="w-6 h-6 text-primary-foreground" />
            </Link>
            <div>
              <h1 className="text-2xl">Mi Progreso</h1>
              <p className="text-sm text-muted-foreground">Seguimiento de tu aprendizaje y logros</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Link
              to="/dashboard"
              className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors"
            >
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
              <div className="w-12 h-12 rounded-lg bg-[#66bb6a] flex items-center justify-center">
                <TrendingUp className="w-6 h-6 text-white" />
              </div>
              <span className="text-3xl">{totalProgress}%</span>
            </div>
            <h3 className="text-sm text-muted-foreground">Progreso Promedio</h3>
          </div>

          <div className="bg-card rounded-lg border border-border p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-lg bg-[#81c784] flex items-center justify-center">
                <Award className="w-6 h-6 text-white" />
              </div>
              <span className="text-3xl">{certificates.length}</span>
            </div>
            <h3 className="text-sm text-muted-foreground">Certificados</h3>
          </div>

          <div className="bg-card rounded-lg border border-border p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-lg bg-[#a5d6a7] flex items-center justify-center">
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
              <h2 className="text-2xl mb-6">Cursos en Progreso</h2>
              <div className="space-y-4">
                {enrolledCourses.map((course) => (
                  <Link
                    key={course.id}
                    to={`/courses/${course.id}`}
                    className="block p-4 rounded-lg border border-border hover:border-primary hover:bg-primary/5 transition-all"
                  >
                    <div className="flex items-start gap-4">
                      <div className="text-4xl flex-shrink-0">{course.image}</div>
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
                      <button className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors flex items-center gap-2">
                        <Download className="w-4 h-4" />
                        Descargar
                      </button>
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

            {/* Learning Streak */}
            <div className="bg-card rounded-lg border border-border p-6">
              <h2 className="text-xl mb-4">Racha de Aprendizaje</h2>
              <div className="text-center mb-4">
                <div className="text-5xl mb-2">🔥</div>
                <p className="text-3xl text-primary mb-1">7 días</p>
                <p className="text-sm text-muted-foreground">¡Sigue así!</p>
              </div>
              <div className="grid grid-cols-7 gap-2">
                {['L', 'M', 'X', 'J', 'V', 'S', 'D'].map((day, index) => (
                  <div
                    key={day}
                    className={`aspect-square rounded-lg flex items-center justify-center text-xs ${
                      index < 7 ? 'bg-primary text-primary-foreground' : 'bg-secondary text-muted-foreground'
                    }`}
                  >
                    {day}
                  </div>
                ))}
              </div>
            </div>

            {/* Study Hours */}
            
          </div>
        </div>
      </main>
    </div>
  );
}
