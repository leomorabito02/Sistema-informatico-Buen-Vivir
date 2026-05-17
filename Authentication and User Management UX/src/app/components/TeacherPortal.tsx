import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Leaf, LogOut, Upload, BookOpen, FileText, Users, BarChart3, CheckCircle, Clock, X, Download } from 'lucide-react';

interface TeacherPortalProps {
  onLogout: () => void;
}

export function TeacherPortal({ onLogout }: TeacherPortalProps) {
  const [showUploadModal, setShowUploadModal] = useState(false);

  const myCourses = [
    {
      id: 1,
      title: 'Panadería Artesanal',
      students: 24,
      materials: 12,
      avgProgress: 68,
      image: '🍞'
    },
    {
      id: 2,
      title: 'Repostería Avanzada',
      students: 18,
      materials: 8,
      avgProgress: 45,
      image: '🎂'
    },
  ];

  const myMaterials = [
    {
      id: 1,
      title: 'Guía de Ingredientes para Panadería',
      course: 'Panadería Artesanal',
      type: 'PDF',
      uploadedDate: '10 de Mayo, 2026',
      status: 'Aprobado',
      downloads: 45
    },
    {
      id: 2,
      title: 'Video: Técnicas de Amasado',
      course: 'Panadería Artesanal',
      type: 'Video',
      uploadedDate: '12 de Mayo, 2026',
      status: 'Aprobado',
      downloads: 38
    },
    {
      id: 3,
      title: 'Recetario de Masas Fermentadas',
      course: 'Panadería Artesanal',
      type: 'PDF',
      uploadedDate: '15 de Mayo, 2026',
      status: 'Pendiente',
      downloads: 0
    },
  ];

  const studentProgress = [
    { name: 'María González', course: 'Panadería Artesanal', progress: 75, lastActivity: 'Hace 2 horas' },
    { name: 'Juan Pérez', course: 'Panadería Artesanal', progress: 82, lastActivity: 'Hace 5 horas' },
    { name: 'Ana Martínez', course: 'Repostería Avanzada', progress: 50, lastActivity: 'Hace 1 día' },
    { name: 'Carlos López', course: 'Panadería Artesanal', progress: 60, lastActivity: 'Hace 3 días' },
  ];

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
              <h1 className="text-2xl">Portal Docente</h1>
              <p className="text-sm text-muted-foreground">Gestiona tus cursos y materiales educativos</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={() => setShowUploadModal(true)}
              className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity flex items-center gap-2"
            >
              <Upload className="w-4 h-4" />
              Subir Material
            </button>
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
        {/* Stats */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div className="bg-card rounded-lg border border-border p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-lg bg-primary flex items-center justify-center">
                <BookOpen className="w-6 h-6 text-primary-foreground" />
              </div>
              <span className="text-3xl">{myCourses.length}</span>
            </div>
            <h3 className="text-sm text-muted-foreground">Mis Cursos</h3>
          </div>

          <div className="bg-card rounded-lg border border-border p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-lg bg-[#66bb6a] flex items-center justify-center">
                <Users className="w-6 h-6 text-white" />
              </div>
              <span className="text-3xl">{myCourses.reduce((sum, c) => sum + c.students, 0)}</span>
            </div>
            <h3 className="text-sm text-muted-foreground">Total Estudiantes</h3>
          </div>

          <div className="bg-card rounded-lg border border-border p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-lg bg-[#81c784] flex items-center justify-center">
                <FileText className="w-6 h-6 text-white" />
              </div>
              <span className="text-3xl">{myMaterials.length}</span>
            </div>
            <h3 className="text-sm text-muted-foreground">Materiales Subidos</h3>
          </div>

          <div className="bg-card rounded-lg border border-border p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-lg bg-[#a5d6a7] flex items-center justify-center">
                <BarChart3 className="w-6 h-6 text-white" />
              </div>
              <span className="text-3xl">
                {Math.round(myCourses.reduce((sum, c) => sum + c.avgProgress, 0) / myCourses.length)}%
              </span>
            </div>
            <h3 className="text-sm text-muted-foreground">Progreso Promedio</h3>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Main Content */}
          <div className="lg:col-span-2 space-y-6">
            {/* My Courses */}
            <div className="bg-card rounded-lg border border-border p-6">
              <h2 className="text-2xl mb-6">Mis Cursos</h2>
              <div className="space-y-4">
                {myCourses.map((course) => (
                  <Link
                    key={course.id}
                    to={`/courses/${course.id}`}
                    className="block p-4 rounded-lg border border-border hover:border-primary hover:bg-primary/5 transition-all"
                  >
                    <div className="flex items-center gap-4">
                      <div className="text-4xl">{course.image}</div>
                      <div className="flex-1">
                        <h3 className="mb-2">{course.title}</h3>
                        <div className="grid grid-cols-3 gap-4 text-sm">
                          <div>
                            <p className="text-muted-foreground">Estudiantes</p>
                            <p className="text-lg">{course.students}</p>
                          </div>
                          <div>
                            <p className="text-muted-foreground">Materiales</p>
                            <p className="text-lg">{course.materials}</p>
                          </div>
                          <div>
                            <p className="text-muted-foreground">Progreso Promedio</p>
                            <p className="text-lg text-primary">{course.avgProgress}%</p>
                          </div>
                        </div>
                      </div>
                    </div>
                  </Link>
                ))}
              </div>
            </div>

            {/* My Materials */}
            <div className="bg-card rounded-lg border border-border p-6">
              <h2 className="text-2xl mb-6">Mis Materiales</h2>
              <div className="space-y-3">
                {myMaterials.map((material) => (
                  <div
                    key={material.id}
                    className="p-4 rounded-lg border border-border"
                  >
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex-1">
                        <h3 className="mb-1">{material.title}</h3>
                        <p className="text-sm text-muted-foreground">
                          {material.course} • {material.type}
                        </p>
                      </div>
                      <span className={`inline-flex px-3 py-1 rounded-full text-xs ${
                        material.status === 'Aprobado'
                          ? 'bg-[#66bb6a]/10 text-[#2d7a3e]'
                          : 'bg-[#f59e0b]/10 text-[#d97706]'
                      }`}>
                        {material.status}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm text-muted-foreground">
                      <span>Subido: {material.uploadedDate}</span>
                      <span>{material.downloads} descargas</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Student Progress */}
            <div className="bg-card rounded-lg border border-border p-6">
              <h2 className="text-xl mb-4">Progreso de Estudiantes</h2>
              <div className="space-y-4">
                {studentProgress.map((student, index) => (
                  <div key={index} className="pb-4 border-b border-border last:border-0 last:pb-0">
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex-1 min-w-0">
                        <p className="mb-1 truncate">{student.name}</p>
                        <p className="text-xs text-muted-foreground">{student.course}</p>
                      </div>
                      <span className="text-lg text-primary ml-2">{student.progress}%</span>
                    </div>
                    <div className="w-full bg-secondary rounded-full h-2 mb-1 overflow-hidden">
                      <div
                        className="h-full bg-primary transition-all"
                        style={{ width: `${student.progress}%` }}
                      />
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Última actividad: {student.lastActivity}
                    </p>
                  </div>
                ))}
              </div>
            </div>

            {/* Content Guidelines */}
            <div className="bg-card rounded-lg border border-border p-6">
              <h2 className="text-xl mb-4">Directrices de Contenido</h2>
              <div className="space-y-3 text-sm">
                <div className="flex items-start gap-2">
                  <CheckCircle className="w-4 h-4 text-primary flex-shrink-0 mt-0.5" />
                  <p>Los materiales deben estar relacionados con el curso</p>
                </div>
                <div className="flex items-start gap-2">
                  <CheckCircle className="w-4 h-4 text-primary flex-shrink-0 mt-0.5" />
                  <p>Formatos aceptados: PDF, DOCX, MP4, imágenes</p>
                </div>
                <div className="flex items-start gap-2">
                  <CheckCircle className="w-4 h-4 text-primary flex-shrink-0 mt-0.5" />
                  <p>Tamaño máximo por archivo: 100 MB</p>
                </div>
                <div className="flex items-start gap-2">
                  <Clock className="w-4 h-4 text-[#f59e0b] flex-shrink-0 mt-0.5" />
                  <p>Los materiales son revisados por administradores antes de publicarse</p>
                </div>
              </div>
            </div>

            {/* Quick Stats */}
            <div className="bg-card rounded-lg border border-border p-6">
              <h2 className="text-xl mb-4">Estadísticas</h2>
              <div className="space-y-3 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Materiales aprobados</span>
                  <span>{myMaterials.filter(m => m.status === 'Aprobado').length}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Materiales pendientes</span>
                  <span>{myMaterials.filter(m => m.status === 'Pendiente').length}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Total descargas</span>
                  <span>{myMaterials.reduce((sum, m) => sum + m.downloads, 0)}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* Upload Material Modal */}
      {showUploadModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-2xl">
            <div className="border-b border-border p-6 flex items-center justify-between">
              <h2 className="text-2xl">Subir Nuevo Material</h2>
              <button
                onClick={() => setShowUploadModal(false)}
                className="p-2 rounded-lg hover:bg-secondary transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6">
              <form className="space-y-4">
                <div>
                  <label className="block mb-2">Título del Material</label>
                  <input
                    type="text"
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="Ej: Guía de Ingredientes"
                  />
                </div>

                <div>
                  <label className="block mb-2">Curso Asociado</label>
                  <select className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary">
                    <option value="">Seleccionar curso</option>
                    {myCourses.map(course => (
                      <option key={course.id} value={course.id}>{course.title}</option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block mb-2">Descripción</label>
                  <textarea
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary min-h-[100px]"
                    placeholder="Describe el contenido del material..."
                  />
                </div>

                <div>
                  <label className="block mb-2">Archivo</label>
                  <div className="border-2 border-dashed border-border rounded-lg p-8 text-center hover:border-primary transition-colors cursor-pointer">
                    <Upload className="w-8 h-8 text-muted-foreground mx-auto mb-2" />
                    <p className="text-sm text-muted-foreground mb-1">Haz clic o arrastra un archivo aquí</p>
                    <p className="text-xs text-muted-foreground">PDF, DOCX, MP4, imágenes (máx. 100MB)</p>
                  </div>
                </div>

                <div className="bg-secondary rounded-lg p-4">
                  <p className="text-sm text-muted-foreground">
                    <strong className="text-foreground">Nota:</strong> Tu material será revisado por un administrador antes de estar disponible para los estudiantes.
                    Recibirás una notificación cuando sea aprobado.
                  </p>
                </div>

                <div className="flex justify-end gap-3 pt-4">
                  <button
                    type="button"
                    onClick={() => setShowUploadModal(false)}
                    className="px-6 py-3 rounded-lg border border-border hover:bg-secondary transition-colors"
                  >
                    Cancelar
                  </button>
                  <button
                    type="submit"
                    className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
                  >
                    Subir Material
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
