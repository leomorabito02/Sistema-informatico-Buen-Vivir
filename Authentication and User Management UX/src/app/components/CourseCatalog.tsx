import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Leaf, Search, Plus, BookOpen, Clock, Users, LogOut, Filter, X, User as UserIcon, FileText, Upload } from 'lucide-react';

interface CourseCatalogProps {
  userRole: 'admin' | 'participante' | 'voluntario' | 'empresa' | 'docente';
  onLogout: () => void;
}

export function CourseCatalog({ userRole, onLogout }: CourseCatalogProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [filterArea, setFilterArea] = useState('all');
  const [showCreateModal, setShowCreateModal] = useState(false);

  const mockCourses = [
    {
      id: 1,
      title: 'Panadería Artesanal',
      description: 'Aprende las técnicas tradicionales y modernas de panadería',
      area: 'Gastronomía',
      teacher: 'María Rodríguez',
      duration: '8 semanas',
      enrolled: 24,
      maxStudents: 30,
      image: '🍞'
    },
    {
      id: 2,
      title: 'Carpintería Básica',
      description: 'Fundamentos de carpintería y trabajo con madera',
      area: 'Oficios',
      teacher: 'Carlos Mendoza',
      duration: '10 semanas',
      enrolled: 18,
      maxStudents: 20,
      image: '🔨'
    },
    {
      id: 3,
      title: 'Costura y Diseño',
      description: 'Técnicas de costura y diseño de prendas',
      area: 'Arte y Diseño',
      teacher: 'Ana López',
      duration: '12 semanas',
      enrolled: 32,
      maxStudents: 35,
      image: '✂️'
    },
    {
      id: 4,
      title: 'Huerta Orgánica',
      description: 'Cultivo de vegetales orgánicos y sustentabilidad',
      area: 'Agricultura',
      teacher: 'Juan Fernández',
      duration: '6 semanas',
      enrolled: 20,
      maxStudents: 25,
      image: '🌱'
    },
    {
      id: 5,
      title: 'Computación Básica',
      description: 'Fundamentos de informática y herramientas digitales',
      area: 'Tecnología',
      teacher: 'Laura Martínez',
      duration: '8 semanas',
      enrolled: 28,
      maxStudents: 30,
      image: '💻'
    },
    {
      id: 6,
      title: 'Cocina Saludable',
      description: 'Preparación de comidas nutritivas y económicas',
      area: 'Gastronomía',
      teacher: 'Roberto García',
      duration: '6 semanas',
      enrolled: 25,
      maxStudents: 30,
      image: '🥗'
    }
  ];

  const areas = ['Gastronomía', 'Oficios', 'Arte y Diseño', 'Agricultura', 'Tecnología'];

  const filteredCourses = mockCourses.filter(course => {
    const matchesSearch = course.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         course.description.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesArea = filterArea === 'all' || course.area === filterArea;
    return matchesSearch && matchesArea;
  });

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
              <h1 className="text-2xl">Catálogo de Cursos</h1>
              <p className="text-sm text-muted-foreground">Explora y gestiona los cursos disponibles</p>
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
        {/* Search and Filters */}
        <div className="bg-card rounded-lg border border-border p-6 mb-6">
          <div className="flex flex-col lg:flex-row gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="Buscar cursos por nombre o descripción..."
              />
            </div>
            <select
              value={filterArea}
              onChange={(e) => setFilterArea(e.target.value)}
              className="px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
            >
              <option value="all">Todas las áreas</option>
              {areas.map(area => (
                <option key={area} value={area}>{area}</option>
              ))}
            </select>
            {userRole === 'admin' && (
              <button
                onClick={() => setShowCreateModal(true)}
                className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity flex items-center gap-2 whitespace-nowrap"
              >
                <Plus className="w-5 h-5" />
                Nuevo Curso
              </button>
            )}
          </div>
        </div>

        {/* Courses Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredCourses.map((course) => (
            <Link
              key={course.id}
              to={`/courses/${course.id}`}
              className="bg-card rounded-lg border border-border overflow-hidden hover:border-primary hover:shadow-lg transition-all group"
            >
              <div className="aspect-video bg-gradient-to-br from-primary to-[#1a3a1f] flex items-center justify-center text-8xl">
                {course.image}
              </div>
              <div className="p-6">
                <div className="flex items-center justify-between mb-2">
                  <span className="inline-flex px-3 py-1 rounded-full text-xs bg-primary/10 text-primary">
                    {course.area}
                  </span>
                  <span className="text-sm text-muted-foreground flex items-center gap-1">
                    <Users className="w-4 h-4" />
                    {course.enrolled}/{course.maxStudents}
                  </span>
                </div>
                <h3 className="text-xl mb-2 group-hover:text-primary transition-colors">{course.title}</h3>
                <p className="text-sm text-muted-foreground mb-4 line-clamp-2">{course.description}</p>
                <div className="flex items-center justify-between text-sm">
                  <div className="flex items-center gap-1 text-muted-foreground">
                    <UserIcon className="w-4 h-4" />
                    {course.teacher}
                  </div>
                  <div className="flex items-center gap-1 text-muted-foreground">
                    <Clock className="w-4 h-4" />
                    {course.duration}
                  </div>
                </div>
                <div className="mt-4 pt-4 border-t border-border">
                  <div className="w-full bg-secondary rounded-full h-2 overflow-hidden">
                    <div
                      className="h-full bg-primary transition-all"
                      style={{ width: `${(course.enrolled / course.maxStudents) * 100}%` }}
                    />
                  </div>
                  <p className="text-xs text-muted-foreground mt-2 text-center">
                    {course.maxStudents - course.enrolled} cupos disponibles
                  </p>
                </div>
              </div>
            </Link>
          ))}
        </div>
      </main>

      {/* Create Course Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-card border-b border-border p-6 flex items-center justify-between">
              <h2 className="text-2xl">Crear Nuevo Curso</h2>
              <button
                onClick={() => setShowCreateModal(false)}
                className="p-2 rounded-lg hover:bg-secondary transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6">
              <form className="space-y-4">
                <div>
                  <label className="block mb-2">Título del Curso</label>
                  <input
                    type="text"
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="Ej: Panadería Artesanal"
                  />
                </div>

                <div>
                  <label className="block mb-2">Descripción</label>
                  <textarea
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary min-h-[100px]"
                    placeholder="Describe el contenido y objetivos del curso..."
                  />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block mb-2">Área Formativa</label>
                    <select className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary">
                      <option value="">Seleccionar área</option>
                      {areas.map(area => (
                        <option key={area} value={area}>{area}</option>
                      ))}
                    </select>
                  </div>

                  <div>
                    <label className="block mb-2">Docente Responsable</label>
                    <select className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary">
                      <option value="">Seleccionar docente</option>
                      <option value="1">María Rodríguez</option>
                      <option value="2">Carlos Mendoza</option>
                      <option value="3">Ana López</option>
                    </select>
                  </div>

                  <div>
                    <label className="block mb-2">Duración</label>
                    <input
                      type="text"
                      className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                      placeholder="Ej: 8 semanas"
                    />
                  </div>

                  <div>
                    <label className="block mb-2">Cupos Máximos</label>
                    <input
                      type="number"
                      className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                      placeholder="30"
                    />
                  </div>
                </div>

                <div>
                  <label className="block mb-2">Materiales Adjuntos</label>
                  <div className="border-2 border-dashed border-border rounded-lg p-8 text-center hover:border-primary transition-colors cursor-pointer">
                    <Upload className="w-8 h-8 text-muted-foreground mx-auto mb-2" />
                    <p className="text-sm text-muted-foreground mb-1">Haz clic o arrastra archivos aquí</p>
                    <p className="text-xs text-muted-foreground">PDF, DOCX, imágenes hasta 10MB cada uno</p>
                  </div>
                </div>

                <div className="bg-secondary rounded-lg p-4">
                  <p className="text-sm text-muted-foreground mb-2">
                    <strong className="text-foreground">Nota:</strong> El contenido del curso se gestiona a través de Contentful (CMS headless).
                    Una vez creado el curso, podrás gestionar las actividades y materiales desde el panel de contenido.
                  </p>
                </div>

                <div className="flex justify-end gap-3 pt-4">
                  <button
                    type="button"
                    onClick={() => setShowCreateModal(false)}
                    className="px-6 py-3 rounded-lg border border-border hover:bg-secondary transition-colors"
                  >
                    Cancelar
                  </button>
                  <button
                    type="submit"
                    className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
                  >
                    Crear Curso
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
