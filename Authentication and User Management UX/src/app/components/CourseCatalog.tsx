import { useState } from 'react';
import { Link } from 'react-router-dom';
import {
  Search, Plus, BookOpen, Clock, Users, LogOut, X, Upload, Edit, AlertCircle, Trash2, Sparkles, Check,
  UtensilsCrossed, Hammer, Scissors, Sprout, Laptop, Zap, Briefcase, Salad, BarChart3, ArrowLeft
} from 'lucide-react';
import { Button } from 'primereact/button';
import { InputText } from 'primereact/inputtext';
import { Dropdown } from 'primereact/dropdown';
import { InputTextarea } from 'primereact/inputtextarea';
import { Toast, ToastState } from './base';

const levelStyles: Record<string, string> = {
  Principiante: 'bg-primary/10 text-primary',
  Intermedio: 'bg-secondary/10 text-secondary',
  Avanzado: 'bg-[#1D1D1B]/10 text-[#1D1D1B]',
};

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
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [courseToDelete, setCourseToDelete] = useState<any>(null);
  const [toast, setToast] = useState<ToastState | null>(null);
  const [acceptedSkills, setAcceptedSkills] = useState<Record<string, boolean>>({});

  const initialCourses = [
    {
      id: 1,
      title: 'Panadería Artesanal',
      description: 'Aprende las técnicas tradicionales y modernas de panadería',
      area: 'Gastronomía',
      teacher: 'María Rodríguez',
      duration: '8 semanas',
      enrolled: 24,
      maxStudents: 30,
      icon: UtensilsCrossed,
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
      icon: Hammer,
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
      icon: Scissors,
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
      icon: Sprout,
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
      icon: Laptop,
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
      icon: Salad,
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
      icon: Zap,
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
      icon: Briefcase,
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

  const [courses, setCourses] = useState(
    initialCourses.map((c) => ({ ...c, status: 'activo' as 'activo' | 'inactivo' }))
  );

  const areas = ['Gastronomía', 'Oficios', 'Arte y Diseño', 'Agricultura', 'Tecnología', 'Administración'];

  const filteredCourses = courses.filter(course => {
    const matchesSearch = course.status === 'activo' && (
      course.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      course.description.toLowerCase().includes(searchQuery.toLowerCase())
    );
    const matchesArea = filterArea === 'all' || course.area === filterArea;
    return matchesSearch && matchesArea;
  });

  const handleCreateSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const data = new FormData(e.currentTarget);
    const title = String(data.get('title') || '').trim();
    if (!title) return;
    setShowCreateModal(false);
    setToast({ type: 'success', message: `El curso "${title}" fue creado correctamente.` });
  };

  const handleEditSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const data = new FormData(e.currentTarget);
    const title = String(data.get('title') || '').trim() || selectedCourseForEdit.title;
    const description = String(data.get('description') || '').trim() || selectedCourseForEdit.description;
    const duration = String(data.get('duration') || '').trim() || selectedCourseForEdit.duration;
    setCourses(courses.map((c) =>
      c.id === selectedCourseForEdit.id ? { ...c, title, description, duration } : c
    ));
    setShowEditModal(false);
    setSelectedCourseForEdit(null);
    setToast({ type: 'success', message: 'Se modificó el curso correctamente' });
  };

  const handleAskDeleteCourse = (course: any) => {
    setCourseToDelete(course);
    setShowDeleteModal(true);
  };

  const confirmDeleteCourse = () => {
    setCourses(courses.map((c) => c.id === courseToDelete.id ? { ...c, status: 'inactivo' } : c));
    setShowDeleteModal(false);
    setToast({ type: 'success', message: 'Se eliminó el curso correctamente' });
    setCourseToDelete(null);
  };

  const skillKey = (courseId: number, skill: string) => `${courseId}::${skill}`;

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
              <h1 className="text-3xl font-bold">Catálogo de Cursos</h1>
              <p className="text-sm text-muted-foreground">Formación práctica para transformar tu futuro</p>
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
            <Button
              onClick={onLogout}
              icon={<LogOut className="w-4 h-4 mr-2" />}
              label="Salir"
              className="px-4 py-2 bg-destructive text-destructive-foreground hover:opacity-90 transition-opacity border-none"
            />
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="p-6 max-w-7xl mx-auto">
        {/* Search and Filters */}
        <div className="bg-card rounded-xl border border-border p-6 mb-8 shadow-sm">
          <div className="flex flex-col lg:flex-row gap-4">
            <div className="flex-1 relative p-input-icon-left">
              <i className="pi pi-search z-10 text-muted-foreground" />
              <InputText
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-3"
                placeholder="Buscar cursos por nombre o descripción..."
              />
            </div>
            <Dropdown
              value={filterArea}
              options={[{label: 'Todas las áreas', value: 'all'}, ...areas.map(a => ({label: a, value: a}))]}
              onChange={(e) => setFilterArea(e.value)}
              className="w-full md:w-auto"
            />
            {userRole === 'admin' && (
              <Button
                onClick={() => setShowCreateModal(true)}
                icon={<Plus className="w-5 h-5 mr-2" />}
                label="Nuevo Curso"
                className="px-6 py-3 bg-primary text-primary-foreground border-none whitespace-nowrap"
              />
            )}
          </div>
        </div>

        {filteredCourses.length === 0 && (
          <div className="text-center py-16 bg-card rounded-xl border border-border border-dashed">
            <BookOpen className="w-10 h-10 text-muted-foreground mx-auto mb-3 opacity-50" />
            <p className="text-muted-foreground">No se encontraron cursos que coincidan con tu búsqueda.</p>
          </div>
        )}

        {/* Courses Grid - Coderhouse style cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
          {filteredCourses.map((course) => (
            <div
              key={course.id}
              className="bg-card rounded-2xl border border-border overflow-hidden hover:shadow-xl hover:-translate-y-1 transition-all duration-200 group h-full flex flex-col relative"
            >
              <Link to={`/courses/${course.id}`} className="flex flex-col flex-1">
                {/* Cover */}
                <div className="relative h-36 bg-gradient-to-br from-primary to-[#1D1D1B] flex items-center justify-center flex-shrink-0">
                  <course.icon className="w-14 h-14 text-white/90" />
                  <span className={`absolute top-3 left-3 px-2.5 py-1 rounded-full text-[11px] font-bold uppercase tracking-wide ${levelStyles[course.level] ?? 'bg-white/20 text-white'}`}>
                    {course.level}
                  </span>
                </div>

                <div className="p-5 flex flex-col flex-1">
                  <span className="inline-flex w-fit px-2.5 py-1 rounded text-xs bg-primary/10 text-primary font-semibold mb-3 uppercase tracking-wide">
                    {course.area}
                  </span>

                  <h3 className="text-lg font-bold mb-1.5 group-hover:text-primary transition-colors leading-snug">
                    {course.title}
                  </h3>
                  <p className="text-sm text-muted-foreground mb-4 line-clamp-2">{course.description}</p>

                  <div className="grid grid-cols-2 gap-2 text-xs text-muted-foreground mb-4">
                    <span className="flex items-center gap-1.5"><Clock className="w-3.5 h-3.5 flex-shrink-0" /> {course.duration}</span>
                    <span className="flex items-center gap-1.5"><BookOpen className="w-3.5 h-3.5 flex-shrink-0" /> {course.modules} módulos</span>
                    <span className="flex items-center gap-1.5 col-span-2 truncate"><BarChart3 className="w-3.5 h-3.5 flex-shrink-0" /> {course.teacher}</span>
                  </div>

                  {/* Suggested Skills*/}
                  {course.suggestedSkills && (
                    <div className="mb-4">
                      <p className="text-[10px] text-muted-foreground mb-1.5 flex items-center gap-1">
                        <Sparkles className="w-3 h-3" /> Sugeridas por IA 
                      </p>
                      <div className="flex flex-wrap gap-1.5">
                        {course.suggestedSkills.slice(0, 2).map((skill: string, idx: number) => {
                          const key = skillKey(course.id, skill);
                          const accepted = acceptedSkills[key];
                          return (
                            <span
                              key={idx}
                              className={`text-[11px] px-2 py-0.5 rounded-full flex items-center gap-1 ${
                                accepted ? 'bg-primary text-primary-foreground' : 'bg-primary/10 text-primary'
                              }`}
                            >
                              {skill}
                              {userRole === 'admin' && !accepted && (
                                <button
                                  type="button"
                                  onClick={(e) => {
                                    e.preventDefault();
                                    setAcceptedSkills((prev) => ({ ...prev, [key]: true }));
                                  }}
                                  className="hover:opacity-70"
                                  title="Aceptar habilidad sugerida"
                                >
                                  <Check className="w-3 h-3" />
                                </button>
                              )}
                            </span>
                          );
                        })}
                      </div>
                    </div>
                  )}

                  {/* Footer: enrollment + CTA */}
                  <div className="mt-auto pt-4 border-t border-border flex items-center justify-between">
                    <div className="flex-1 mr-3">
                      <div className="flex items-center justify-between mb-1">
                        <span className="text-xs font-medium flex items-center gap-1"><Users className="w-3 h-3" /> {course.enrolled}/{course.maxStudents}</span>
                      </div>
                      <div className="w-full bg-muted rounded-full h-1.5 overflow-hidden">
                        <div
                          className="h-full bg-secondary transition-all"
                          style={{ width: `${(course.enrolled / course.maxStudents) * 100}%` }}
                        />
                      </div>
                    </div>
                    <span className="text-sm font-semibold text-primary whitespace-nowrap group-hover:underline">
                      Ver curso →
                    </span>
                  </div>
                </div>
              </Link>

              {/* Edit Button - visible for admin and voluntario */}
              {(userRole === 'admin' || userRole === 'voluntario') && (
                  <Button
                    onClick={(e) => {
                      e.preventDefault();
                      setSelectedCourseForEdit(course);
                      setShowEditModal(true);
                    }}
                    icon={<Edit className="w-4 h-4 text-primary" />}
                    className="absolute top-3 right-3 p-button-rounded p-button-text bg-white/90 hover:bg-white shadow-md opacity-0 group-hover:opacity-100 transition-all"
                    tooltip="Editar curso"
                  />
              )}

              {/* Delete Button - visible for admin only */}
              {userRole === 'admin' && (
                  <Button
                    onClick={(e) => {
                      e.preventDefault();
                      handleAskDeleteCourse(course);
                    }}
                    icon={<Trash2 className="w-4 h-4 text-destructive" />}
                    className="absolute top-3 right-14 p-button-rounded p-button-text bg-white/90 hover:bg-white shadow-md opacity-0 group-hover:opacity-100 transition-all"
                    tooltip="Eliminar curso"
                  />
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
              <form className="space-y-4" onSubmit={handleCreateSubmit}>
                <div>
                  <label className="block mb-2">Título del Curso</label>
                  <InputText
                    name="title"
                    className="w-full"
                    placeholder="Ej: Panadería Artesanal"
                    required
                  />
                </div>

                <div>
                  <label className="block mb-2">Descripción</label>
                  <InputTextarea
                    name="description"
                    className="w-full resize-none min-h-[100px]"
                    placeholder="Describe el contenido y objetivos del curso..."
                    required
                  />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block mb-2">Área Formativa</label>
                    <Dropdown options={areas.map(a => ({label: a, value: a}))} placeholder="Seleccionar área" className="w-full" />
                  </div>

                  <div>
                    <label className="block mb-2">Docente Responsable</label>
                    <Dropdown
                      options={[
                        {label: 'María Rodríguez', value: '1'},
                        {label: 'Carlos Mendoza', value: '2'},
                        {label: 'Ana López', value: '3'}
                      ]}
                      placeholder="Seleccionar docente"
                      className="w-full"
                    />
                  </div>

                  <div>
                    <label className="block mb-2">Duración</label>
                    <InputText
                      name="duration"
                      className="w-full"
                      placeholder="Ej: 8 semanas"
                    />
                  </div>

                  <div>
                    <label className="block mb-2">Cupos Máximos</label>
                    <InputText
                      type="number"
                      className="w-full"
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

                <div className="bg-white border border-border rounded-lg p-4">
                  <p className="text-sm text-muted-foreground mb-2">
                    <strong className="text-foreground">Nota:</strong> El contenido del curso se gestiona a través de Contentful (CMS headless).
                    Una vez creado el curso, podrás gestionar las actividades y materiales desde el panel de contenido.
                  </p>
                </div>

                <div className="flex justify-end gap-3 pt-4">
                  <Button
                    type="button"
                    onClick={() => setShowCreateModal(false)}
                    label="Cancelar"
                    className="p-button-outlined p-button-secondary"
                  />
                  <Button
                    type="submit"
                    label="Crear Curso"
                    className="bg-primary text-primary-foreground border-none"
                  />
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
              <form className="space-y-4" onSubmit={handleEditSubmit}>
                <div>
                  <label className="block mb-2">Título del Curso</label>
                  <InputText
                    name="title"
                    defaultValue={selectedCourseForEdit.title}
                    className="w-full"
                    placeholder="Ej: Panadería Artesanal"
                  />
                </div>

                <div>
                  <label className="block mb-2">Descripción</label>
                  <InputTextarea
                    name="description"
                    defaultValue={selectedCourseForEdit.description}
                    className="w-full resize-none min-h-[100px]"
                    placeholder="Describe el contenido y objetivos del curso..."
                  />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block mb-2">Área Formativa</label>
                    <Dropdown 
                      value={selectedCourseForEdit.area}
                      options={areas.map(a => ({label: a, value: a}))} 
                      placeholder="Seleccionar área" 
                      className="w-full" 
                    />
                  </div>

                  <div>
                    <label className="block mb-2">Docente Responsable</label>
                    <Dropdown 
                      value={selectedCourseForEdit.teacher}
                      options={[
                        {label: 'María Rodríguez', value: 'María Rodríguez'},
                        {label: 'Carlos Mendoza', value: 'Carlos Mendoza'},
                        {label: 'Ana López', value: 'Ana López'}
                      ]} 
                      placeholder="Seleccionar docente" 
                      className="w-full" 
                    />
                  </div>

                  <div>
                    <label className="block mb-2">Duración</label>
                    <InputText
                      name="duration"
                      defaultValue={selectedCourseForEdit.duration}
                      className="w-full"
                      placeholder="Ej: 8 semanas"
                    />
                  </div>

                  <div>
                    <label className="block mb-2">Cupos Máximos</label>
                    <InputText
                      type="number"
                      defaultValue={selectedCourseForEdit.maxStudents}
                      className="w-full"
                      placeholder="30"
                    />
                  </div>
                </div>

                <div className="bg-primary/5 border border-primary/20 rounded-lg p-4">
                  <div className="flex items-start gap-3">
                    <AlertCircle className="w-5 h-5 text-primary mt-0.5 flex-shrink-0" />
                    <p className="text-sm text-foreground">
                      <strong>Información:</strong> Los cambios en este formulario actualizarán los metadatos del curso.
                      El contenido del curso se gestiona a través de Contentful (CMS headless).
                    </p>
                  </div>
                </div>

                <div className="flex justify-end gap-3 pt-4">
                  <Button
                    type="button"
                    onClick={() => setShowEditModal(false)}
                    label="Cancelar"
                    className="p-button-outlined p-button-secondary"
                  />
                  <Button
                    type="submit"
                    label="Guardar Cambios"
                    className="bg-primary text-primary-foreground border-none"
                  />
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {showDeleteModal && courseToDelete && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-sm p-6">
            <h2 className="text-xl font-semibold mb-3">Eliminar Curso</h2>
            <p className="text-sm text-muted-foreground mb-2">
              ¿Confirmás que querés eliminar <strong className="text-foreground">{courseToDelete.title}</strong>?
            </p>
            {courseToDelete.enrolled > 0 && (
              <div className="flex items-start gap-2 text-sm text-secondary bg-secondary/10 border border-secondary/20 rounded-lg p-3 mb-4">
                <AlertCircle className="w-4 h-4 flex-shrink-0 mt-0.5" />
                <span>Este curso tiene {courseToDelete.enrolled} participantes inscriptos activos. Dejará de verse en el catálogo, pero ellos conservarán el curso en su historial.</span>
              </div>
            )}
            <div className="flex gap-3 mt-4">
              <button
                onClick={() => setShowDeleteModal(false)}
                className="flex-1 px-4 py-2.5 rounded-lg border border-border hover:bg-secondary transition-colors"
              >
                Cancelar
              </button>
              <button
                onClick={confirmDeleteCourse}
                className="flex-1 px-4 py-2.5 bg-destructive text-destructive-foreground rounded-lg hover:opacity-90 transition-opacity"
              >
                Eliminar
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
