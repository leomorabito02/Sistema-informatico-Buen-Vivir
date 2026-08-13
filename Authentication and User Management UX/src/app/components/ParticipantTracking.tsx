import { useState } from 'react';
import { Link } from 'react-router-dom';
import { LogOut, Search, Plus, User, Calendar, TrendingUp, FileText, Download, X, Edit, AlertCircle, ArrowLeft } from 'lucide-react';
import { Toast, ToastState } from './base';

interface ParticipantTrackingProps {
  userRole: 'admin' | 'voluntario';
  onLogout: () => void;
}

const emptyFicha = { nombre: '', apellido: '', edad: '', skills: '', currentSituation: '', progressNotes: '' };

export function ParticipantTracking({ userRole, onLogout }: ParticipantTrackingProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [showMilestoneModal, setShowMilestoneModal] = useState(false);
  const [selectedParticipant, setSelectedParticipant] = useState<any>(null);
  const [showTimelineModal, setShowTimelineModal] = useState(false);
  const [editingMilestone, setEditingMilestone] = useState<any>(null);
  const [showFichaModal, setShowFichaModal] = useState(false);
  const [fichaMode, setFichaMode] = useState<'create' | 'edit'>('create');
  const [fichaForm, setFichaForm] = useState(emptyFicha);
  const [fichaError, setFichaError] = useState('');
  const [toast, setToast] = useState<ToastState | null>(null);

  const [participants, setParticipants] = useState([
    {
      id: 1,
      nombre: 'María',
      apellido: 'González',
      age: 28,
      currentSituation: 'Buscando empleo activamente',
      progressNotes: 'Completó Panadería Artesanal, en búsqueda activa de empleo en el rubro.',
      skills: ['Panadería', 'Trabajo en equipo', 'Puntualidad'],
      lastUpdate: '2 días atrás',
      milestones: 5,
      progress: 75
    },
    {
      id: 2,
      nombre: 'Juan',
      apellido: 'Pérez',
      age: 35,
      currentSituation: 'En proceso de formación',
      progressNotes: 'Cursando Carpintería Básica, buen desempeño en las actividades prácticas.',
      skills: ['Carpintería básica', 'Responsabilidad'],
      lastUpdate: '1 semana atrás',
      milestones: 3,
      progress: 50
    },
    {
      id: 3,
      nombre: 'Ana',
      apellido: 'Martínez',
      age: 24,
      currentSituation: 'Cursando taller de costura',
      progressNotes: 'Avance destacado en Costura y Diseño, próxima a finalizar el módulo 5.',
      skills: ['Costura', 'Diseño', 'Creatividad'],
      lastUpdate: '3 días atrás',
      milestones: 7,
      progress: 85
    },
  ]);

  const milestones = [
    {
      id: 1,
      date: '15 de Mayo, 2026',
      title: 'Completó curso de Panadería Artesanal',
      description: 'Finalizó con éxito el curso de panadería, demostrando gran compromiso y habilidad',
      responsible: 'Laura Martínez'
    },
    {
      id: 2,
      date: '10 de Mayo, 2026',
      title: 'Primera entrevista laboral',
      description: 'Participó en su primera entrevista laboral en Panadería Don José',
      responsible: 'Carlos López'
    },
    {
      id: 3,
      date: '5 de Mayo, 2026',
      title: 'Inscripción a curso de computación',
      description: 'Se inscribió al curso de computación básica para ampliar sus habilidades',
      responsible: 'Ana Rodríguez'
    },
  ];

  const filteredParticipants = participants.filter(p =>
    `${p.nombre} ${p.apellido}`.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleViewTimeline = (participant: any) => {
    setSelectedParticipant(participant);
    setShowTimelineModal(true);
  };

  const handleAddMilestone = (participant: any) => {
    setSelectedParticipant(participant);
    setShowMilestoneModal(true);
  };

  const handleNewFicha = () => {
    setFichaMode('create');
    setSelectedParticipant(null);
    setFichaForm(emptyFicha);
    setFichaError('');
    setShowFichaModal(true);
  };

  const handleEditFicha = (participant: any) => {
    setFichaMode('edit');
    setSelectedParticipant(participant);
    setFichaForm({
      nombre: participant.nombre,
      apellido: participant.apellido,
      edad: String(participant.age),
      skills: participant.skills.join(', '),
      currentSituation: participant.currentSituation,
      progressNotes: participant.progressNotes,
    });
    setFichaError('');
    setShowFichaModal(true);
  };

  const handleFichaSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const { nombre, apellido, edad, skills, currentSituation, progressNotes } = fichaForm;
    if (!nombre.trim() || !apellido.trim() || !edad.trim() || !skills.trim() || !currentSituation.trim() || !progressNotes.trim()) {
      setFichaError('Completar todos los campos obligatorios para continuar');
      return;
    }

    if (fichaMode === 'create') {
      const newParticipant = {
        id: Date.now(),
        nombre: nombre.trim(),
        apellido: apellido.trim(),
        age: Number(edad),
        currentSituation: currentSituation.trim(),
        progressNotes: progressNotes.trim(),
        skills: skills.split(',').map(s => s.trim()).filter(Boolean),
        lastUpdate: 'Recién creada',
        milestones: 0,
        progress: 0,
      };
      setParticipants([newParticipant, ...participants]);
      setToast({ type: 'success', message: 'Ficha personal creada correctamente' });
    } else {
      setParticipants(participants.map(p => p.id === selectedParticipant.id ? {
        ...p,
        nombre: nombre.trim(),
        apellido: apellido.trim(),
        age: Number(edad),
        currentSituation: currentSituation.trim(),
        progressNotes: progressNotes.trim(),
        skills: skills.split(',').map(s => s.trim()).filter(Boolean),
      } : p));
      setToast({ type: 'success', message: 'Ficha personal modificada correctamente' });
    }

    setFichaError('');
    setShowFichaModal(false);
    setSelectedParticipant(null);
  };

  const handleExportPdf = (participant: any) => {
    setToast({ type: 'success', message: `Ficha de ${participant.nombre} ${participant.apellido} exportada en PDF.` });
  };

  return (
    <div className="min-h-screen bg-background">
      <Toast toast={toast} onClose={() => setToast(null)} />

      {/* Header */}
      <header className="bg-primary/90 text-white border-b-2 border-primary p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link to="/dashboard" className="w-10 h-10 rounded-lg bg-white/20 flex items-center justify-center">
              <img src="/logo.png" alt="Buen Vivir" className="w-8 h-8 object-contain" />
            </Link>
            <div>
              <h1 className="text-2xl text-white">Seguimiento de Participantes</h1>
              <p className="text-sm text-white/70">Gestión de fichas y procesos personales</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Link
              to="/dashboard"
              className="px-4 py-2 rounded-lg border border-white hover:bg-white/10 transition-colors text-white flex items-center gap-2"
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
        {/* Search */}
        <div className="bg-card rounded-lg border border-border p-6 mb-6">
          <div className="flex flex-col lg:flex-row gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="Buscar participantes por nombre..."
              />
            </div>
            <button
              onClick={handleNewFicha}
              className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity flex items-center gap-2 whitespace-nowrap"
            >
              <Plus className="w-5 h-5" />
              Nueva Ficha
            </button>
          </div>
        </div>

        {/* Participants List */}
        <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
          {filteredParticipants.map((participant) => (
            <div key={participant.id} className="bg-card rounded-lg border border-border p-6">
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center gap-3">
                  <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center">
                    <User className="w-6 h-6 text-primary" />
                  </div>
                  <div>
                    <h3 className="mb-1">{participant.nombre} {participant.apellido}</h3>
                    <p className="text-sm text-muted-foreground">{participant.age} años</p>
                  </div>
                </div>
                <div className="flex items-start gap-2">
                  <button
                    onClick={() => handleEditFicha(participant)}
                    className="p-1 rounded hover:bg-secondary transition-colors flex-shrink-0"
                    title="Editar ficha personal"
                  >
                    <Edit className="w-4 h-4 text-muted-foreground" />
                  </button>
                </div>
              </div>

              <div className="mb-4">
                <p className="text-sm text-muted-foreground mb-1">Situación actual:</p>
                <p className="text-sm">{participant.currentSituation}</p>
              </div>

              <div className="mb-4">
                <p className="text-sm text-muted-foreground mb-2">Habilidades:</p>
                <div className="flex flex-wrap gap-2">
                  {participant.skills.map((skill, index) => (
                    <span
                      key={index}
                      className="inline-flex px-2 py-1 rounded-full text-xs bg-primary/10 text-primary"
                    >
                      {skill}
                    </span>
                  ))}
                </div>
              </div>

              <div className="flex items-center justify-between text-sm text-muted-foreground mb-4">
              </div>

              <div className="flex gap-2">
                <button
                  onClick={() => handleViewTimeline(participant)}
                  className="flex-1 px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors flex items-center justify-center gap-2"
                >
                  <Calendar className="w-4 h-4" />
                  Ver Timeline
                </button>
                <button
                  onClick={() => handleAddMilestone(participant)}
                  className="flex-1 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity flex items-center justify-center gap-2"
                >
                  <Plus className="w-4 h-4" />
                  Nuevo Hito
                </button>
              </div>
            </div>
          ))}
        </div>

        {filteredParticipants.length === 0 && (
          <div className="text-center py-12">
            <p className="text-muted-foreground text-lg">No se encontraron participantes</p>
          </div>
        )}
      </main>

      {/* Timeline Modal */}
      {showTimelineModal && selectedParticipant && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-3xl max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-card border-b border-border p-6 flex items-center justify-between">
              <div>
                <h2 className="text-2xl mb-1">Línea de Tiempo - {selectedParticipant.nombre} {selectedParticipant.apellido}</h2>
                <p className="text-sm text-muted-foreground">Historial completo de avances y hitos</p>
              </div>
              <button
                onClick={() => setShowTimelineModal(false)}
                className="p-2 rounded-lg hover:bg-secondary transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6">
              <div className="flex items-center justify-between mb-6">
                <button
                  onClick={() => handleExportPdf(selectedParticipant)}
                  className="px-6 py-3 rounded-lg border border-border hover:bg-secondary transition-colors flex items-center gap-2"
                >
                  <Download className="w-4 h-4" />
                  Exportar PDF
                </button>
                <Link
                  to="/asistencia-general"
                  className="px-6 py-3 rounded-lg border border-border hover:bg-secondary transition-colors flex items-center gap-2 text-sm"
                >
                  Ver Registro de Asistencia General
                </Link>
              </div>

              <div className="space-y-6">
                {milestones.map((milestone, index) => (
                  <div key={milestone.id} className="flex gap-4">
                    <div className="flex flex-col items-center">
                      <div className="w-10 h-10 rounded-full bg-primary flex items-center justify-center text-primary-foreground flex-shrink-0">
                        <TrendingUp className="w-5 h-5" />
                      </div>
                      {index < milestones.length - 1 && (
                        <div className="w-0.5 flex-1 bg-border mt-2" />
                      )}
                    </div>
                    <div className="flex-1 pb-6">
                      <div className="flex items-start justify-between mb-2">
                        <h3 className="text-lg">{milestone.title}</h3>
                        <div className="flex items-center gap-2 ml-4">
                          <button
                            onClick={() => {
                              setEditingMilestone(milestone);
                              setShowMilestoneModal(true);
                            }}
                            className="p-1 rounded hover:bg-secondary transition-colors"
                          >
                            <Edit className="w-4 h-4" />
                          </button>
                          <span className="text-sm text-muted-foreground whitespace-nowrap">{milestone.date}</span>
                        </div>
                      </div>
                      <p className="text-muted-foreground mb-2">{milestone.description}</p>
                      <p className="text-sm text-muted-foreground">Responsable: {milestone.responsible}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Add/Edit Milestone Modal */}
      {showMilestoneModal && selectedParticipant && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-2xl">
            <div className="border-b border-border p-6 flex items-center justify-between">
              <h2 className="text-2xl">
                {editingMilestone ? 'Editar Hito' : 'Registrar Nuevo Hito'} - {selectedParticipant.nombre} {selectedParticipant.apellido}
              </h2>
              <button
                onClick={() => {
                  setShowMilestoneModal(false);
                  setEditingMilestone(null);
                }}
                className="p-2 rounded-lg hover:bg-secondary transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6">
              <form className="space-y-4">
                <div>
                  <label className="block mb-2">Fecha del Hito</label>
                  <input
                    type="date"
                    defaultValue={editingMilestone ? editingMilestone.date : ''}
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    required
                  />
                  <p className="text-xs text-muted-foreground mt-1">Formato: Día/Mes/Año</p>
                </div>

                <div>
                  <label className="block mb-2">Título del Hito</label>
                  <input
                    type="text"
                    defaultValue={editingMilestone ? editingMilestone.title : ''}
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="Ej: Completó curso de Panadería"
                    required
                  />
                </div>

                <div>
                  <label className="block mb-2">Descripción</label>
                  <textarea
                    defaultValue={editingMilestone ? editingMilestone.description : ''}
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary min-h-[120px]"
                    placeholder="Describe el logro o avance significativo..."
                    required
                  />
                </div>

                <div>
                  <label className="block mb-2">Responsable</label>
                  <input
                    type="text"
                    defaultValue={editingMilestone ? editingMilestone.responsible : ''}
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="Tu nombre"
                    required
                  />
                </div>

                <div className="flex justify-end gap-3 pt-4">
                  <button
                    type="button"
                    onClick={() => {
                      setShowMilestoneModal(false);
                      setEditingMilestone(null);
                    }}
                    className="px-6 py-3 rounded-lg border border-border hover:bg-secondary transition-colors"
                  >
                    Cancelar
                  </button>
                  <button
                    type="submit"
                    className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
                  >
                    {editingMilestone ? 'Guardar Cambios' : 'Registrar Hito'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Create/Edit Ficha Personal Modal (HU-29/30/31) */}
      {showFichaModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-card border-b border-border p-6 flex items-center justify-between">
              <h2 className="text-2xl">
                {fichaMode === 'create' ? 'Nueva Ficha de Proceso Personal' : 'Editar Ficha de Proceso Personal'}
              </h2>
              <button
                onClick={() => setShowFichaModal(false)}
                className="p-2 rounded-lg hover:bg-secondary transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6">
              <form onSubmit={handleFichaSubmit} className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block mb-2">Nombre *</label>
                    <input
                      type="text"
                      value={fichaForm.nombre}
                      onChange={(e) => setFichaForm({ ...fichaForm, nombre: e.target.value })}
                      className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                      placeholder="María"
                    />
                  </div>
                  <div>
                    <label className="block mb-2">Apellido *</label>
                    <input
                      type="text"
                      value={fichaForm.apellido}
                      onChange={(e) => setFichaForm({ ...fichaForm, apellido: e.target.value })}
                      className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                      placeholder="González"
                    />
                  </div>
                  <div>
                    <label className="block mb-2">Edad *</label>
                    <input
                      type="number"
                      value={fichaForm.edad}
                      onChange={(e) => setFichaForm({ ...fichaForm, edad: e.target.value })}
                      className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                      placeholder="28"
                    />
                  </div>
                  <div>
                    <label className="block mb-2">Habilidades * (separadas por coma)</label>
                    <input
                      type="text"
                      value={fichaForm.skills}
                      onChange={(e) => setFichaForm({ ...fichaForm, skills: e.target.value })}
                      className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                      placeholder="Panadería, Trabajo en equipo"
                    />
                  </div>
                </div>

                <div>
                  <label className="block mb-2">Situación Personal *</label>
                  <textarea
                    value={fichaForm.currentSituation}
                    onChange={(e) => setFichaForm({ ...fichaForm, currentSituation: e.target.value })}
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary min-h-[80px]"
                    placeholder="Describe la situación actual del participante..."
                  />
                </div>

                <div>
                  <label className="block mb-2">Avances *</label>
                  <textarea
                    value={fichaForm.progressNotes}
                    onChange={(e) => setFichaForm({ ...fichaForm, progressNotes: e.target.value })}
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary min-h-[80px]"
                    placeholder="Describe los avances y logros del participante..."
                  />
                </div>

                {fichaError && (
                  <div className="flex items-start gap-2 text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-lg p-3">
                    <AlertCircle className="w-4 h-4 flex-shrink-0 mt-0.5" />
                    <span>{fichaError}</span>
                  </div>
                )}

                <div className="flex justify-end gap-3 pt-4">
                  <button
                    type="button"
                    onClick={() => setShowFichaModal(false)}
                    className="px-6 py-3 rounded-lg border border-border hover:bg-secondary transition-colors"
                  >
                    Cancelar
                  </button>
                  <button
                    type="submit"
                    className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
                  >
                    {fichaMode === 'create' ? 'Crear Ficha' : 'Guardar Cambios'}
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
