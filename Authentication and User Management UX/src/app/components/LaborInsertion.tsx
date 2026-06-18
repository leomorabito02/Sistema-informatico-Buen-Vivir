import { useState } from 'react';
import { Link } from 'react-router-dom';
import { LogOut, Briefcase, Users, Plus, Edit, Trash2, Eye, MapPin, DollarSign, Clock, AlertCircle, Monitor } from 'lucide-react';
import { Dialog } from 'primereact/dialog';
import { InputText } from 'primereact/inputtext';
import { Dropdown } from 'primereact/dropdown';
import { InputTextarea } from 'primereact/inputtextarea';
import { Button } from 'primereact/button';
import { Sidebar } from 'primereact/sidebar';

interface LaborInsertionProps {
  userRole: 'admin' | 'voluntario' | 'empresa';
  onLogout: () => void;
}

export function LaborInsertion({ userRole, onLogout }: LaborInsertionProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [showJobModal, setShowJobModal] = useState(false);
  const [showCandidateSidebar, setShowCandidateSidebar] = useState(false);
  const [selectedJob, setSelectedJob] = useState<any>(null);

  const [newJob, setNewJob] = useState({
    title: '',
    company: userRole === 'empresa' ? 'Mi Empresa Mock' : '',
    location: '',
    salary: '',
    type: 'Empleo',
    description: '',
    requirements: '',
    schedule: '',
    modality: 'Presencial'
  });

  const [jobs, setJobs] = useState([
    {
      id: 1,
      title: 'Panadero/a Artesanal',
      company: 'Panadería La Tradición',
      location: 'Alta Gracia',
      salary: '$25.000 - $30.000',
      description: 'Buscamos panadero/a con experiencia en productos artesanales',
      requirements: ['Experiencia en panadería', 'Disponibilidad full-time', 'Responsabilidad'],
      candidates: [1, 2],
      status: 'Activa',
      type: 'Empleo',
      schedule: 'Lunes a Viernes 06:00 a 14:00',
      modality: 'Presencial',
      postDate: '2026-05-20',
    },
    {
      id: 2,
      title: 'Jardinero',
      company: 'Mi Empresa Mock',
      location: 'Alta Gracia',
      salary: '$40.000 - $50.000',
      description: 'Se busca jardinero para mantenimiento de áreas verdes.',
      requirements: ['Plantación', 'Cuidado de plantas'],
      candidates: [3],
      status: 'Activa',
      type: 'Empleo',
      schedule: 'Flexible 8 horas',
      modality: 'Presencial',
      postDate: '2026-05-18',
    },
    {
      id: 3,
      title: 'Atención al Cliente',
      company: 'Mi Empresa Mock',
      location: 'Alta Gracia',
      salary: '$20.000 - $25.000',
      description: 'Atención presencial en sucursal centro.',
      requirements: ['Secundario completo', 'Buena presencia', 'Proactividad'],
      candidates: [1, 3],
      status: 'Activa',
      type: 'Empleo',
      schedule: 'Lunes a Sábados 09:00 a 13:00',
      modality: 'Presencial',
      postDate: '2026-05-15',
    }
  ]);

  const allCandidates = [
    {
      id: 1,
      name: 'Juan Pérez',
      courseCompleted: 'Panadería Artesanal',
      skills: ['Técnicas de amasado', 'Atención al cliente'],
      status: 'Disponible',
      completionDate: '2026-04-15',
    },
    {
      id: 2,
      name: 'María García',
      courseCompleted: 'Costura y Diseño',
      skills: ['Máquina de coser', 'Responsabilidad'],
      status: 'En Entrevista',
      completionDate: '2026-03-20',
    },
    {
      id: 3,
      name: 'Carlos López',
      courseCompleted: 'Cuidado de plantas y jardinería',
      skills: ['Plantación', 'Cuidado de plantas'],
      status: 'Disponible',
      completionDate: '2026-02-10',
    },
  ];

  const handleCreateJob = () => {
    const job = {
      id: Date.now(),
      title: newJob.title,
      company: newJob.company,
      location: newJob.location,
      salary: newJob.salary,
      description: newJob.description,
      requirements: newJob.requirements.split(',').map(r => r.trim()).filter(r => r),
      candidates: [],
      status: 'Activa',
      type: newJob.type,
      schedule: newJob.schedule,
      modality: newJob.modality,
      postDate: new Date().toISOString().split('T')[0]
    };
    setJobs([job, ...jobs]);
    setShowJobModal(false);
    setNewJob({
      title: '',
      company: userRole === 'empresa' ? 'Mi Empresa Mock' : '',
      location: '',
      salary: '',
      type: 'Empleo',
      description: '',
      requirements: '',
      schedule: '',
      modality: 'Presencial'
    });
  };

  const visibleJobs = jobs.filter(job => {
    const matchesSearch = job.title.toLowerCase().includes(searchQuery.toLowerCase()) || 
                          job.company.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesRole = userRole === 'empresa' ? job.company === 'Mi Empresa Mock' : true;
    return matchesSearch && matchesRole;
  });

  const jobCandidates = selectedJob ? allCandidates.filter(c => selectedJob.candidates.includes(c.id)) : [];
  const totalCandidatesForVisible = visibleJobs.reduce((acc, job) => acc + job.candidates.length, 0);

  return (
    <div className="min-h-screen bg-background flex flex-col">
      {/* Header */}
      <header className="bg-card border-b border-border p-6 shadow-sm">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link to="/dashboard" className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center shadow-md">
              <img src="/logo.png" alt="Buen Vivir" className="w-8 h-8 object-contain" />
            </Link>
            <div>
              <h1 className="text-2xl font-bold text-foreground">Gestión de Ofertas Laborales</h1>
              <p className="text-sm text-muted-foreground font-medium">Panel de {userRole === 'empresa' ? 'Empresa' : 'Administración'}</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Link
              to="/dashboard"
              className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors font-medium"
            >
              Volver al Dashboard
            </Link>
            <Button
              onClick={onLogout}
              icon={<LogOut className="w-4 h-4 mr-2" />}
              label="Salir"
              className="px-4 py-2 bg-destructive text-destructive-foreground hover:opacity-90 border-none font-medium"
            />
          </div>
        </div>
      </header>

      <main className="flex-1 p-6 max-w-7xl mx-auto w-full">
        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <div className="bg-card rounded-xl border border-border p-6 shadow-sm">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center">
                <Briefcase className="w-6 h-6 text-primary" />
              </div>
              <span className="text-4xl font-bold text-foreground">{visibleJobs.length}</span>
            </div>
            <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">Tus Ofertas Activas</h3>
          </div>

          <div className="bg-card rounded-xl border border-border p-6 shadow-sm">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-lg bg-blue-500/10 flex items-center justify-center">
                <Users className="w-6 h-6 text-blue-500" />
              </div>
              <span className="text-4xl font-bold text-foreground">{totalCandidatesForVisible}</span>
            </div>
            <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">Total Postulantes</h3>
          </div>

          <div className="bg-card rounded-xl border border-border p-6 shadow-sm">
            <div className="flex items-center justify-between mb-4">
              <div className="w-12 h-12 rounded-lg bg-green-500/10 flex items-center justify-center">
                <AlertCircle className="w-6 h-6 text-green-500" />
              </div>
              <span className="text-4xl font-bold text-foreground">0</span>
            </div>
            <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">Contrataciones Exitosas</h3>
          </div>
        </div>

        {/* Main Content */}
        <div className="bg-card rounded-xl border border-border p-6 shadow-sm">
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between mb-8 gap-4 border-b border-border pb-6">
            <h2 className="text-2xl font-bold text-foreground">Listado de Ofertas</h2>
            <div className="flex items-center gap-3 w-full sm:w-auto">
              <span className="p-input-icon-left w-full sm:w-72">
                <i className="pi pi-search" />
                <InputText
                  placeholder="Buscar por puesto..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full"
                />
              </span>
              <Button
                onClick={() => setShowJobModal(true)}
                icon={<Plus className="w-5 h-5 mr-2" />}
                label="Nueva Oferta"
                className="bg-primary text-primary-foreground border-none px-5 py-2 font-semibold shadow-md hover:opacity-90"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {visibleJobs.length === 0 ? (
              <div className="col-span-full py-12 text-center text-muted-foreground">
                <Briefcase className="w-12 h-12 mx-auto mb-4 opacity-50" />
                <p className="text-lg">No se encontraron ofertas laborales.</p>
              </div>
            ) : (
              visibleJobs.map(job => (
                <div key={job.id} className="bg-background rounded-xl border border-border p-6 hover:border-primary/50 transition-all hover:shadow-md flex flex-col h-full">
                  <div className="mb-4">
                    <span className={`inline-block px-3 py-1 rounded-full text-xs font-bold mb-3 ${job.type === 'Empleo' ? 'bg-blue-100 text-blue-700' : 'bg-purple-100 text-purple-700'}`}>
                      {job.type}
                    </span>
                    <h3 className="text-xl font-bold text-foreground leading-tight">{job.title}</h3>
                    {userRole !== 'empresa' && <p className="text-sm text-primary font-semibold mt-1">{job.company}</p>}
                  </div>

                  <div className="space-y-3 mb-6 flex-1">
                    <div className="flex items-center gap-3 text-sm text-muted-foreground">
                      <div className="w-8 h-8 rounded-full bg-secondary flex items-center justify-center flex-shrink-0">
                        <MapPin className="w-4 h-4 text-foreground" />
                      </div>
                      <span className="font-medium">{job.location}</span>
                    </div>
                    <div className="flex items-center gap-3 text-sm text-muted-foreground">
                      <div className="w-8 h-8 rounded-full bg-secondary flex items-center justify-center flex-shrink-0">
                        <Monitor className="w-4 h-4 text-foreground" />
                      </div>
                      <span className="font-medium">{job.modality}</span>
                    </div>
                    <div className="flex items-center gap-3 text-sm text-muted-foreground">
                      <div className="w-8 h-8 rounded-full bg-secondary flex items-center justify-center flex-shrink-0">
                        <Clock className="w-4 h-4 text-foreground" />
                      </div>
                      <span className="font-medium">{job.schedule}</span>
                    </div>
                    <div className="flex items-center gap-3 text-sm font-medium text-foreground">
                      <div className="w-8 h-8 rounded-full bg-green-100 flex items-center justify-center flex-shrink-0">
                        <DollarSign className="w-4 h-4 text-green-700" />
                      </div>
                      <span>{job.salary}</span>
                    </div>
                  </div>

                  <div className="pt-4 border-t border-border mt-auto flex items-center justify-between">
                    <div className="flex items-center gap-2 text-sm font-bold text-primary bg-primary/10 px-3 py-1.5 rounded-lg">
                      <Users className="w-4 h-4" />
                      <span>{job.candidates.length} Postulantes</span>
                    </div>
                    <Button 
                      onClick={() => {
                        setSelectedJob(job);
                        setShowCandidateSidebar(true);
                      }}
                      label="Ver Detalle" 
                      className="p-button-text p-button-sm font-bold p-0 text-foreground hover:text-primary transition-colors hover:bg-transparent" 
                    />
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </main>

      {/* Candidate Sidebar */}
      <Sidebar 
        visible={showCandidateSidebar} 
        position="right" 
        onHide={() => setShowCandidateSidebar(false)}
        className="w-full md:w-96 lg:w-[32rem]"
      >
        {selectedJob && (
          <div className="flex flex-col h-full">
            <div className="mb-6 pb-6 border-b border-border">
              <h2 className="text-2xl font-bold mb-2">Candidatos Postulados</h2>
              <p className="text-muted-foreground text-sm">
                Puesto: <span className="font-bold text-foreground text-base ml-1">{selectedJob.title}</span>
              </p>
            </div>
            
            <div className="flex-1 overflow-y-auto pr-2">
              {jobCandidates.length === 0 ? (
                <div className="text-center p-8 bg-muted/30 rounded-xl border border-dashed border-border mt-10">
                  <Users className="w-16 h-16 text-muted-foreground/30 mx-auto mb-4" />
                  <p className="text-lg font-medium text-muted-foreground">Aún no hay postulantes</p>
                  <p className="text-sm text-muted-foreground/70 mt-2">Los candidatos que apliquen aparecerán aquí.</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {jobCandidates.map(candidate => (
                    <div key={candidate.id} className="bg-card rounded-xl border border-border p-5 shadow-sm hover:shadow-md transition-shadow">
                      <div className="flex items-start justify-between mb-4">
                        <div className="flex items-center gap-3">
                          <div className="w-10 h-10 rounded-full bg-primary/20 flex items-center justify-center font-bold text-primary">
                            {candidate.name.charAt(0)}
                          </div>
                          <div>
                            <h3 className="font-bold text-lg leading-tight">{candidate.name}</h3>
                            <p className="text-sm text-primary font-medium">{candidate.courseCompleted}</p>
                          </div>
                        </div>
                        <span className={`px-2 py-1 rounded text-xs font-bold ${
                          candidate.status === 'Disponible' ? 'bg-green-100 text-green-700' :
                          'bg-yellow-100 text-yellow-700'
                        }`}>
                          {candidate.status}
                        </span>
                      </div>

                      <div className="mb-5">
                        <p className="text-xs text-muted-foreground mb-2 font-bold uppercase tracking-wider">Habilidades destacadas</p>
                        <div className="flex flex-wrap gap-2">
                          {candidate.skills.map((skill, idx) => (
                            <span key={idx} className="text-xs px-3 py-1 rounded-full bg-secondary text-secondary-foreground font-semibold">
                              {skill}
                            </span>
                          ))}
                        </div>
                      </div>

                      <Button label="Ver Perfil Completo" className="w-full p-button-outlined font-semibold" />
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}
      </Sidebar>

      {/* New Job Modal */}
      <Dialog 
        header={<span className="text-2xl font-bold">Publicar Nueva Oferta</span>} 
        visible={showJobModal} 
        style={{ width: '60vw' }} 
        breakpoints={{ '960px': '80vw', '641px': '95vw' }}
        onHide={() => setShowJobModal(false)}
        className="font-sans"
      >
        <div className="space-y-6 pt-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="flex flex-col gap-2">
              <label className="font-bold text-sm text-muted-foreground uppercase">Puesto a cubrir</label>
              <InputText 
                placeholder="Ej: Desarrollador, Ayudante, etc." 
                value={newJob.title}
                onChange={(e) => setNewJob({...newJob, title: e.target.value})}
                className="w-full p-3" 
              />
            </div>
            <div className="flex flex-col gap-2">
              <label className="font-bold text-sm text-muted-foreground uppercase">Empresa</label>
              <InputText 
                placeholder="Nombre de tu empresa" 
                value={newJob.company}
                onChange={(e) => setNewJob({...newJob, company: e.target.value})}
                disabled={userRole === 'empresa'}
                className="w-full p-3" 
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="flex flex-col gap-2">
              <label className="font-bold text-sm text-muted-foreground uppercase">Ubicación</label>
              <InputText 
                placeholder="Ej: Alta Gracia o Remoto" 
                value={newJob.location}
                onChange={(e) => setNewJob({...newJob, location: e.target.value})}
                className="w-full p-3" 
              />
            </div>
            <div className="flex flex-col gap-2">
              <label className="font-bold text-sm text-muted-foreground uppercase">Modalidad</label>
              <Dropdown 
                options={['Presencial', 'Híbrido', 'Remoto']} 
                value={newJob.modality}
                onChange={(e) => setNewJob({...newJob, modality: e.value})}
                className="w-full" 
              />
            </div>
            <div className="flex flex-col gap-2">
              <label className="font-bold text-sm text-muted-foreground uppercase">Tipo</label>
              <Dropdown 
                options={['Empleo', 'Pasantía']} 
                value={newJob.type}
                onChange={(e) => setNewJob({...newJob, type: e.value})}
                className="w-full" 
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="flex flex-col gap-2">
              <label className="font-bold text-sm text-muted-foreground uppercase">Horario</label>
              <InputText 
                placeholder="Ej: Lunes a Viernes 9 a 18hs" 
                value={newJob.schedule}
                onChange={(e) => setNewJob({...newJob, schedule: e.target.value})}
                className="w-full p-3" 
              />
            </div>
            <div className="flex flex-col gap-2">
              <label className="font-bold text-sm text-muted-foreground uppercase">Salario (Opcional)</label>
              <InputText 
                placeholder="Ej: $450.000 o A convenir" 
                value={newJob.salary}
                onChange={(e) => setNewJob({...newJob, salary: e.target.value})}
                className="w-full p-3" 
              />
            </div>
          </div>

          <div className="flex flex-col gap-2">
            <label className="font-bold text-sm text-muted-foreground uppercase">Descripción del rol</label>
            <InputTextarea 
              rows={4} 
              placeholder="Describe detalladamente qué tareas realizará, el ambiente de trabajo, etc." 
              value={newJob.description}
              onChange={(e) => setNewJob({...newJob, description: e.target.value})}
              className="w-full p-3 resize-none" 
            />
          </div>

          <div className="flex flex-col gap-2">
            <label className="font-bold text-sm text-muted-foreground uppercase">Requisitos (separados por coma)</label>
            <InputText 
              placeholder="Ej: Secundario completo, Proactividad, Excel" 
              value={newJob.requirements}
              onChange={(e) => setNewJob({...newJob, requirements: e.target.value})}
              className="w-full p-3" 
            />
          </div>

          <div className="flex justify-end gap-3 pt-6 border-t border-border mt-4">
            <Button 
              label="Cancelar" 
              icon="pi pi-times" 
              outlined 
              severity="secondary" 
              onClick={() => setShowJobModal(false)} 
              className="px-6 font-bold"
            />
            <Button 
              label="Publicar Oferta" 
              icon="pi pi-check" 
              onClick={handleCreateJob} 
              disabled={!newJob.title || !newJob.location}
              className="px-6 font-bold bg-primary text-primary-foreground border-none"
            />
          </div>
        </div>
      </Dialog>
    </div>
  );
}
