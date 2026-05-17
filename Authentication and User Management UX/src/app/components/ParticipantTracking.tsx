import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Leaf, LogOut, Search, Plus, User, Calendar, TrendingUp, FileText, Download, X } from 'lucide-react';

interface ParticipantTrackingProps {
  userRole: 'admin' | 'voluntario';
  onLogout: () => void;
}

export function ParticipantTracking({ userRole, onLogout }: ParticipantTrackingProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [showMilestoneModal, setShowMilestoneModal] = useState(false);
  const [selectedParticipant, setSelectedParticipant] = useState<any>(null);
  const [showTimelineModal, setShowTimelineModal] = useState(false);

  const participants = [
    {
      id: 1,
      name: 'María González',
      age: 28,
      currentSituation: 'Buscando empleo activamente',
      skills: ['Panadería', 'Trabajo en equipo', 'Puntualidad'],
      lastUpdate: '2 días atrás',
      milestones: 5,
      progress: 75
    },
    {
      id: 2,
      name: 'Juan Pérez',
      age: 35,
      currentSituation: 'En proceso de formación',
      skills: ['Carpintería básica', 'Responsabilidad'],
      lastUpdate: '1 semana atrás',
      milestones: 3,
      progress: 50
    },
    {
      id: 3,
      name: 'Ana Martínez',
      age: 24,
      currentSituation: 'Cursando taller de costura',
      skills: ['Costura', 'Diseño', 'Creatividad'],
      lastUpdate: '3 días atrás',
      milestones: 7,
      progress: 85
    },
  ];

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
    p.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleViewTimeline = (participant: any) => {
    setSelectedParticipant(participant);
    setShowTimelineModal(true);
  };

  const handleAddMilestone = (participant: any) => {
    setSelectedParticipant(participant);
    setShowMilestoneModal(true);
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
              <h1 className="text-2xl">Seguimiento de Participantes</h1>
              <p className="text-sm text-muted-foreground">Gestión de fichas y procesos personales</p>
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
        {/* Search */}
        <div className="bg-card rounded-lg border border-border p-6 mb-6">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
              placeholder="Buscar participantes por nombre..."
            />
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
                    <h3 className="mb-1">{participant.name}</h3>
                    <p className="text-sm text-muted-foreground">{participant.age} años</p>
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-2xl text-primary mb-1">{participant.progress}%</div>
                  <p className="text-xs text-muted-foreground">Progreso</p>
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
                <span>{participant.milestones} hitos registrados</span>
                <span>Actualizado {participant.lastUpdate}</span>
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
                <h2 className="text-2xl mb-1">Línea de Tiempo - {selectedParticipant.name}</h2>
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
                <div className="text-center flex-1">
                  <div className="text-3xl text-primary mb-1">{selectedParticipant.progress}%</div>
                  <p className="text-sm text-muted-foreground">Progreso General</p>
                </div>
                <div className="text-center flex-1">
                  <div className="text-3xl text-primary mb-1">{selectedParticipant.milestones}</div>
                  <p className="text-sm text-muted-foreground">Hitos Registrados</p>
                </div>
                <button className="px-6 py-3 rounded-lg border border-border hover:bg-secondary transition-colors flex items-center gap-2">
                  <Download className="w-4 h-4" />
                  Exportar PDF
                </button>
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
                        <span className="text-sm text-muted-foreground whitespace-nowrap ml-4">{milestone.date}</span>
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

      {/* Add Milestone Modal */}
      {showMilestoneModal && selectedParticipant && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-2xl">
            <div className="border-b border-border p-6 flex items-center justify-between">
              <h2 className="text-2xl">Registrar Nuevo Hito - {selectedParticipant.name}</h2>
              <button
                onClick={() => setShowMilestoneModal(false)}
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
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    required
                  />
                </div>

                <div>
                  <label className="block mb-2">Título del Hito</label>
                  <input
                    type="text"
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="Ej: Completó curso de Panadería"
                    required
                  />
                </div>

                <div>
                  <label className="block mb-2">Descripción</label>
                  <textarea
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary min-h-[120px]"
                    placeholder="Describe el logro o avance significativo..."
                    required
                  />
                </div>

                <div>
                  <label className="block mb-2">Responsable</label>
                  <input
                    type="text"
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="Tu nombre"
                    required
                  />
                </div>

                <div className="flex justify-end gap-3 pt-4">
                  <button
                    type="button"
                    onClick={() => setShowMilestoneModal(false)}
                    className="px-6 py-3 rounded-lg border border-border hover:bg-secondary transition-colors"
                  >
                    Cancelar
                  </button>
                  <button
                    type="submit"
                    className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
                  >
                    Registrar Hito
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
