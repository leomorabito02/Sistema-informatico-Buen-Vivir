import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Leaf, Search, Plus, BookOpen, Clock, Users, LogOut, Filter, X, User as UserIcon, FileText, Upload, Edit, AlertCircle } from 'lucide-react';

interface CourseCatalogProps {
  userRole: 'admin' | 'participante' | 'voluntario' | 'empresa' | 'docente';
  onLogout: () => void;
}

export function CourseCatalog({ userRole, onLogout }: CourseCatalogProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [filterArea, setFilterArea] = useState('all');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [selectedCourseForEdit, setSelectedCourseForEdit] = useState<any>(null);

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
      image: '🍞',
      level: 'Principiante',
      schedule: 'Martes y Jueves, 18:00-20:00',
      startDate: '2026-06-01',
      rating: 4.8,
      reviews: 45,
      modules: 12,
      completedModules: 9,
      suggestedSkills: ['Técnicas de amasado', 'Fermentación', 'Horneado profesional'],
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
      image: '🔨',
      level: 'Principiante',
      schedule: 'Lunes y Miércoles, 14:00-16:00',
      startDate: '2026-06-05',
      rating: 4.6,
      reviews: 32,
      modules: 15,
      completedModules: 0,
      suggestedSkills: ['Manejo de herramientas', 'Seguridad', 'Medición y corte'],
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
      image: '✂️',
      level: 'Intermedio',
      schedule: 'Miércoles y Viernes, 16:00-18:00',
      startDate: '2026-06-10',
      rating: 4.9,
      reviews: 58,
      modules: 16,
      completedModules: 5,
      suggestedSkills: ['Máquina de coser', 'Diseño de prendas', 'Telas'],
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
      image: '🌱',
      level: 'Principiante',
      schedule: 'Sábados, 10:00-12:00',
      startDate: '2026-06-07',
      rating: 4.7,
      reviews: 41,
      modules: 8,
      completedModules: 3,
      suggestedSkills: ['Cultivo orgánico', 'Compostaje', 'Riego eficiente'],
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
      image: '💻',
      level: 'Principiante',
      schedule: 'Martes y Jueves, 19:00-21:00',
      startDate: '2026-06-03',
      rating: 4.5,
      reviews: 67,
      modules: 10,
      completedModules: 7,
      suggestedSkills: ['Sistema operativo', 'Navegación web', 'Ofimática'],
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
      image: '🥗',
      level: 'Principiante',
      schedule: 'Lunes y Miércoles, 17:00-19:00',
      startDate: '2026-06-02',
      rating: 4.7,
      reviews: 52,
      modules: 8,
      completedModules: 0,
      suggestedSkills: ['Nutrición', 'Cocina saludable', 'Presupuesto'],
    },
    {
      id: 7,
      title: 'Electricidad Residencial',
      description: 'Instalaciones eléctricas seguras para viviendas',
      area: 'Oficios',
      teacher: 'Miguel Torres',
      duration: '10 semanas',
      enrolled: 15,
      maxStudents: 18,
      image: '⚡',
      level: 'Avanzado',
      schedule: 'Jueves y Sábado, 09:00-11:00',
      startDate: '2026-06-12',
      rating: 4.8,
      reviews: 28,
      modules: 14,
      completedModules: 0,
      suggestedSkills: ['Seguridad eléctrica', 'Circuitos', 'Herramientas'],
    },
    {
      id: 8,
      title: 'Gestión de Negocios',
      description: 'Principios de administración y emprendimiento',
      area: 'Administración',
      teacher: 'Patricia González',
      duration: '8 semanas',
      enrolled: 22,
      maxStudents: 28,
      image: '💼',
      level: 'Intermedio',
      schedule: 'Martes y Viernes, 17:00-19:00',
      startDate: '2026-06-15',
      rating: 4.6,
      reviews: 39,
      modules: 11,
      completedModules: 0,
      suggestedSkills: ['Plan de negocios', 'Finanzas', 'Liderazgo'],
    }
  ];

  const areas = ['Gastronomía', 'Oficios', 'Arte y Diseño', 'Agricultura', 'Tecnología', 'Administración'];

  const filteredCourses = mockCourses.filter(course => {
    const matchesSearch = course.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         course.description.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesArea = filterArea === 'all' || course.area === filterArea;
    return matchesSearch && matchesArea;
  });

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="bg-primary/90 text-white border-b-2 border-primary p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link to="/dashboard" className="w-10 h-10 rounded-lg bg-white/20 flex items-center justify-center">
              <Leaf className="w-6 h-6 text-white" />
            </Link>
            <div>
              <h1 className="text-2xl text-white">Catálogo de Cursos</h1>
              <p className="text-sm text-white/70">Explora y gestiona los cursos disponibles</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Link
              to="/dashboard"
              className="px-4 py-2 rounded-lg border border-white hover:bg-white/10 transition-colors text-white"
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
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {filteredCourses.map((course) => (
            <div
              key={course.id}
              className="bg-card rounded-lg border border-border overflow-hidden hover:border-primary hover:shadow-lg transition-all group h-full flex flex-col relative"
            >
              <Link
                to={`/courses/${course.id}`}
                className="flex flex-col flex-1"
              >
                <div className="aspect-square bg-gradient-to-br from-primary to-[#1a3a1f] flex items-center justify-center text-5xl flex-shrink-0">
                  {course.image}
                </div>
                <div className="p-3 flex flex-col flex-1">
                  <div className="flex items-center justify-between mb-2 gap-1">
                    <span className="inline-flex px-2 py-0.5 rounded text-xs bg-primary/10 text-primary font-medium truncate">
                      {course.area}
                    </span>
                  </div>
                  <h3 className="text-sm font-semibold mb-1 group-hover:text-primary transition-colors line-clamp-2">{course.title}</h3>
                  <p className="text-xs text-muted-foreground mb-2 line-clamp-1">{course.description}</p>
                  
                  {/* Suggested Skills */}
                  {course.suggestedSkills && (
                    <div className="mb-2">
                      <p className="text-xs font-medium text-muted-foreground mb-1">Habilidades sugeridas:</p>
                      <div className="flex flex-wrap gap-1">
                        {course.suggestedSkills.map((skill: string, idx: number) => (
                          <span key={idx} className="text-xs px-2 py-0.5 rounded bg-green-100 text-green-700">
                            {skill}
                          </span>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Progress Bar */}
                  <div className="mt-auto pt-2 border-t border-gray-200">
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-xs font-medium">{course.enrolled}/{course.maxStudents}</span>
                      <span className="text-xs text-muted-foreground">{Math.round((course.enrolled / course.maxStudents) * 100)}%</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-1.5 overflow-hidden">
                      <div
                        className="h-full bg-primary transition-all"
                        style={{ width: `${(course.enrolled / course.maxStudents) * 100}%` }}
                      />
                    </div>
                  </div>
                </div>
              </Link>
              
              {/* Edit Button - visible for admin and voluntario */}
              {(userRole === 'admin' || userRole === 'voluntario') && (
                <button
                  onClick={() => {
                    setSelectedCourseForEdit(course);
                    setShowEditModal(true);
                  }}
                  className="absolute top-3 right-3 p-2 bg-white/90 rounded-lg hover:bg-white shadow-md transition-all opacity-0 group-hover:opacity-100"
                  title="Editar curso"
                >
                  <Edit className="w-4 h-4 text-primary" />
                </button>
              )}
            </div>
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

      {/* Edit Course Modal */}
      {showEditModal && selectedCourseForEdit && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-card border-b border-border p-6 flex items-center justify-between">
              <h2 className="text-2xl">Editar Curso</h2>
              <button
                onClick={() => setShowEditModal(false)}
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
                    defaultValue={selectedCourseForEdit.title}
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="Ej: Panadería Artesanal"
                  />
                </div>

                <div>
                  <label className="block mb-2">Descripción</label>
                  <textarea
                    defaultValue={selectedCourseForEdit.description}
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary min-h-[100px]"
                    placeholder="Describe el contenido y objetivos del curso..."
                  />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block mb-2">Área Formativa</label>
                    <select defaultValue={selectedCourseForEdit.area} className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary">
                      <option value="">Seleccionar área</option>
                      {areas.map(area => (
                        <option key={area} value={area}>{area}</option>
                      ))}
                    </select>
                  </div>

                  <div>
                    <label className="block mb-2">Docente Responsable</label>
                    <select defaultValue={selectedCourseForEdit.teacher} className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary">
                      <option value="">Seleccionar docente</option>
                      <option value="María Rodríguez">María Rodríguez</option>
                      <option value="Carlos Mendoza">Carlos Mendoza</option>
                      <option value="Ana López">Ana López</option>
                    </select>
                  </div>

                  <div>
                    <label className="block mb-2">Duración</label>
                    <input
                      type="text"
                      defaultValue={selectedCourseForEdit.duration}
                      className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                      placeholder="Ej: 8 semanas"
                    />
                  </div>

                  <div>
                    <label className="block mb-2">Cupos Máximos</label>
                    <input
                      type="number"
                      defaultValue={selectedCourseForEdit.maxStudents}
                      className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                      placeholder="30"
                    />
                  </div>
                </div>

                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                  <div className="flex items-start gap-3">
                    <AlertCircle className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0" />
                    <p className="text-sm text-blue-800">
                      <strong>Información:</strong> Los cambios en este formulario actualizarán los metadatos del curso. 
                      El contenido del curso se gestiona a través de Contentful (CMS headless).
                    </p>
                  </div>
                </div>

                <div className="flex justify-end gap-3 pt-4">
                  <button
                    type="button"
                    onClick={() => setShowEditModal(false)}
                    className="px-6 py-3 rounded-lg border border-border hover:bg-secondary transition-colors"
                  >
                    Cancelar
                  </button>
                  <button
                    type="submit"
                    className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
                  >
                    Guardar Cambios
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
