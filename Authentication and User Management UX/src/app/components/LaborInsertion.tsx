import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Leaf, LogOut, Briefcase, Users, Search, Plus, Edit, Trash2, Eye, X, MapPin, DollarSign, Clock, AlertCircle } from 'lucide-react';

interface LaborInsertionProps {
  userRole: 'admin' | 'voluntario' | 'empresa';
  onLogout: () => void;
}

export function LaborInsertion({ userRole, onLogout }: LaborInsertionProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [showJobModal, setShowJobModal] = useState(false);
  const [showCandidateModal, setShowCandidateModal] = useState(false);
  const [selectedJob, setSelectedJob] = useState<any>(null);

  const mockJobs = [
    {
      id: 1,
      title: 'Panadero/a Artesanal',
      company: 'Panadería La Tradición',
      location: 'Alta Gracia',
      salary: '$25.000 - $30.000',
      description: 'Buscamos panadero/a con experiencia en productos artesanales',
      requirements: ['Experiencia en panadería', 'Disponibilidad full-time', 'Responsabilidad'],
      candidates: 3,
      status: 'Activa',
      type: 'Empleo',
      postDate: '2026-05-20',
    },
    {
      id: 2,
      title: 'Carpintero/a Ayudante',
      company: 'Construcciones García',
      location: 'Alta Gracia',
      salary: '$18.000 - $22.000',
      description: 'Pasantía en carpintería básica con posibilidad de contratación',
      requirements: ['Disponibilidad de aprendizaje', 'Responsabilidad', 'Capacidad física'],
      candidates: 5,
      status: 'Activa',
      type: 'Pasantía',
      postDate: '2026-05-18',
    },
    {
      id: 3,
      title: 'Costurero/a',
      company: 'Taller de Costura Estilo',
      location: 'La Cumbre',
      salary: '$20.000 - $25.000',
      description: 'Posición en taller de costura para productos artesanales',
      requirements: ['Manejo de máquina de coser', 'Experiencia en diseño', 'Atención al detalle'],
      candidates: 2,
      status: 'Activa',
      type: 'Empleo',
      postDate: '2026-05-15',
    },
    {
      id: 4,
      title: 'Oficial Electricista',
      company: 'Instalaciones Eléctricas RJ',
      location: 'Alta Gracia',
      salary: '$28.000 - $35.000',
      description: 'Oficial electricista con experiencia en instalaciones residenciales',
      requirements: ['Certificación en electricidad', 'Experiencia 2+ años', 'Conocimiento de seguridad'],
      candidates: 1,
      status: 'Activa',
      type: 'Empleo',
      postDate: '2026-05-12',
    },
  ];

  const mockCandidates = [
    {
      id: 1,
      name: 'Juan Pérez',
      courseCompleted: 'Panadería Artesanal',
      skills: ['Técnicas de amasado', 'Fermentación', 'Horneado'],
      status: 'Disponible',
      appliedJobs: 2,
      completionDate: '2026-04-15',
    },
    {
      id: 2,
      name: 'María García',
      courseCompleted: 'Costura y Diseño',
      skills: ['Máquina de coser', 'Diseño de prendas', 'Telas'],
      status: 'En Entrevista',
      appliedJobs: 3,
      completionDate: '2026-03-20',
    },
    {
      id: 3,
      name: 'Carlos López',
      courseCompleted: 'Carpintería Básica',
      skills: ['Medición', 'Herramientas', 'Seguridad'],
      status: 'Contratado',
      appliedJobs: 1,
      completionDate: '2026-02-10',
    },
  ];

  const filteredJobs = mockJobs.filter(job =>
    job.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
    job.company.toLowerCase().includes(searchQuery.toLowerCase())
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
              <h1 className="text-2xl">Inserción Laboral</h1>
              <p className="text-sm text-muted-foreground">Gestión de empleos, pasantías y candidatos</p>
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

      <main className="p-6">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
          {/* Stats */}
          <div className="bg-card rounded-lg border border-border p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-lg bg-primary flex items-center justify-center">
                <Briefcase className="w-6 h-6 text-primary-foreground" />
              </div>
              <span className="text-3xl">{mockJobs.length}</span>
            </div>
            <h3 className="text-sm text-muted-foreground">Ofertas Activas</h3>
          </div>

          <div className="bg-card rounded-lg border border-border p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-lg bg-green-500 flex items-center justify-center">
                <Users className="w-6 h-6 text-white" />
              </div>
              <span className="text-3xl">{mockCandidates.length}</span>
            </div>
            <h3 className="text-sm text-muted-foreground">Candidatos Disponibles</h3>
          </div>

          <div className="bg-card rounded-lg border border-border p-6">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-lg bg-blue-500 flex items-center justify-center">
                <AlertCircle className="w-6 h-6 text-white" />
              </div>
              <span className="text-3xl">{mockCandidates.filter(c => c.status === 'Contratado').length}</span>
            </div>
            <h3 className="text-sm text-muted-foreground">Contrataciones</h3>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Job Listings */}
          <div className="lg:col-span-2 space-y-6">
            <div className="bg-card rounded-lg border border-border p-6">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-2xl">Ofertas de Empleo y Pasantías</h2>
                {(userRole === 'admin' || userRole === 'empresa') && (
                  <button
                    onClick={() => setShowJobModal(true)}
                    className="px-4 py-2 rounded-lg bg-primary text-primary-foreground hover:opacity-90 transition-opacity flex items-center gap-2"
                  >
                    <Plus className="w-4 h-4" />
                    Nueva Oferta
                  </button>
                )}
              </div>

              {/* Search */}
              <div className="mb-4 flex items-center gap-2 bg-input-background rounded-lg p-3 border border-border">
                <Search className="w-4 h-4 text-muted-foreground" />
                <input
                  type="text"
                  placeholder="Buscar por título o empresa..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="flex-1 bg-transparent outline-none"
                />
              </div>

              {/* Jobs List */}
              <div className="space-y-4">
                {filteredJobs.map(job => (
                  <div key={job.id} className="bg-background rounded-lg border border-border p-4 hover:border-primary transition-colors">
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex-1">
                        <h3 className="text-lg font-semibold">{job.title}</h3>
                        <p className="text-sm text-muted-foreground">{job.company}</p>
                      </div>
                      <span className={`px-3 py-1 rounded text-xs font-medium ${job.type === 'Empleo' ? 'bg-blue-100 text-blue-700' : 'bg-green-100 text-green-700'}`}>
                        {job.type}
                      </span>
                    </div>

                    <p className="text-sm text-muted-foreground mb-3">{job.description}</p>

                    <div className="grid grid-cols-3 gap-4 mb-3 text-sm">
                      <div className="flex items-center gap-2 text-muted-foreground">
                        <MapPin className="w-4 h-4" />
                        {job.location}
                      </div>
                      <div className="flex items-center gap-2 text-muted-foreground">
                        <DollarSign className="w-4 h-4" />
                        {job.salary}
                      </div>
                      <div className="flex items-center gap-2 text-muted-foreground">
                        <Users className="w-4 h-4" />
                        {job.candidates} candidatos
                      </div>
                    </div>

                    <div className="flex items-center gap-2 pt-3 border-t border-border">
                      {(userRole === 'admin' || userRole === 'empresa') && (
                        <>
                          <button className="px-3 py-1 rounded text-sm border border-border hover:bg-secondary transition-colors flex items-center gap-1">
                            <Edit className="w-4 h-4" />
                            Editar
                          </button>
                          <button className="px-3 py-1 rounded text-sm border border-border hover:bg-destructive/10 transition-colors flex items-center gap-1 text-destructive">
                            <Trash2 className="w-4 h-4" />
                            Eliminar
                          </button>
                        </>
                      )}
                      <button className="px-3 py-1 rounded text-sm bg-primary text-primary-foreground hover:opacity-90 transition-opacity flex items-center gap-1 ml-auto">
                        <Eye className="w-4 h-4" />
                        Ver Candidatos
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Candidates Sidebar */}
          <div className="bg-card rounded-lg border border-border p-6">
            <h2 className="text-xl mb-4">Candidatos Disponibles</h2>

            <div className="space-y-4">
              {mockCandidates.map(candidate => (
                <div key={candidate.id} className="bg-background rounded-lg border border-border p-4">
                  <div className="flex items-start justify-between mb-2">
                    <div>
                      <h3 className="font-semibold text-sm">{candidate.name}</h3>
                      <p className="text-xs text-muted-foreground">{candidate.courseCompleted}</p>
                    </div>
                    <span className={`px-2 py-1 rounded text-xs font-medium ${
                      candidate.status === 'Disponible' ? 'bg-green-100 text-green-700' :
                      candidate.status === 'En Entrevista' ? 'bg-yellow-100 text-yellow-700' :
                      'bg-blue-100 text-blue-700'
                    }`}>
                      {candidate.status}
                    </span>
                  </div>

                  <div className="mb-2">
                    <p className="text-xs text-muted-foreground mb-1">Habilidades:</p>
                    <div className="flex flex-wrap gap-1">
                      {candidate.skills.map((skill, idx) => (
                        <span key={idx} className="text-xs px-2 py-0.5 rounded bg-primary/10 text-primary">
                          {skill}
                        </span>
                      ))}
                    </div>
                  </div>

                  <div className="flex items-center gap-2 text-xs text-muted-foreground mb-3 pt-2 border-t border-border">
                    <Clock className="w-3 h-3" />
                    {candidate.appliedJobs} solicitudes
                  </div>

                  <button className="w-full px-3 py-1.5 rounded text-sm bg-primary text-primary-foreground hover:opacity-90 transition-opacity">
                    Ver Perfil
                  </button>
                </div>
              ))}
            </div>
          </div>
        </div>
      </main>

      {/* Modals */}
      {showJobModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-2xl p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-2xl">Nueva Oferta de Empleo</h2>
              <button
                onClick={() => setShowJobModal(false)}
                className="p-2 rounded-lg hover:bg-secondary transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <form className="space-y-4">
              <div>
                <label className="block mb-2 font-medium">Título del Puesto</label>
                <input
                  type="text"
                  className="w-full px-4 py-2 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                  placeholder="Ej: Panadero/a Artesanal"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block mb-2 font-medium">Empresa</label>
                  <input
                    type="text"
                    className="w-full px-4 py-2 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="Nombre de la empresa"
                  />
                </div>

                <div>
                  <label className="block mb-2 font-medium">Ubicación</label>
                  <input
                    type="text"
                    className="w-full px-4 py-2 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="Ciudad/Zona"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block mb-2 font-medium">Salario</label>
                  <input
                    type="text"
                    className="w-full px-4 py-2 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="$20.000 - $25.000"
                  />
                </div>

                <div>
                  <label className="block mb-2 font-medium">Tipo</label>
                  <select className="w-full px-4 py-2 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary">
                    <option>Empleo</option>
                    <option>Pasantía</option>
                  </select>
                </div>
              </div>

              <div>
                <label className="block mb-2 font-medium">Descripción</label>
                <textarea
                  className="w-full px-4 py-2 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary min-h-[100px]"
                  placeholder="Describe la posición..."
                />
              </div>

              <div className="flex gap-2 justify-end pt-4 border-t border-border">
                <button
                  type="button"
                  onClick={() => setShowJobModal(false)}
                  className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors"
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-lg bg-primary text-primary-foreground hover:opacity-90 transition-opacity"
                >
                  Crear Oferta
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
