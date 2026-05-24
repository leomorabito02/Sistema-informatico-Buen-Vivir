import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Leaf, LogOut, Plus, Search, Edit, Trash2, Eye, X, Upload, Calendar } from 'lucide-react';

interface CampaignManagementProps {
  userRole: 'admin' | 'voluntario';
  onLogout: () => void;
}

export function CampaignManagement({ userRole, onLogout }: CampaignManagementProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [modalMode, setModalMode] = useState<'create' | 'edit'>('create');
  const [selectedCampaign, setSelectedCampaign] = useState<any>(null);

  const campaigns = [
    {
      id: 1,
      title: 'Día Mundial del Medio Ambiente',
      description: 'Jornada de concientización sobre el cuidado del medio ambiente',
      date: '5 de Junio, 2026',
      category: 'Ambiental',
      status: 'Publicada',
      views: 234,
      createdBy: 'Laura Martínez'
    },
    {
      id: 2,
      title: 'Taller de Compostaje Comunitario',
      description: 'Aprende a hacer compost en casa y reduce tus residuos',
      date: '15 de Mayo, 2026',
      category: 'Educación Ambiental',
      status: 'Publicada',
      views: 156,
      createdBy: 'Carlos López'
    },
    {
      id: 3,
      title: 'Alimentación Saludable',
      description: 'Serie de charlas sobre nutrición y hábitos alimenticios',
      date: '20 de Mayo, 2026',
      category: 'Salud',
      status: 'Borrador',
      views: 0,
      createdBy: 'Ana Rodríguez'
    },
    {
      id: 4,
      title: 'Huerta Urbana Ecológica',
      description: 'Crea tu propia huerta sustentable en espacios reducidos',
      date: '12 de Junio, 2026',
      category: 'Ecología',
      status: 'Publicada',
      views: 189,
      createdBy: 'Roberto García'
    },
    {
      id: 5,
      title: 'Reciclaje y Economía Circular',
      description: 'Aprende técnicas de reciclaje y reutilización de materiales',
      date: '18 de Junio, 2026',
      category: 'Ecología',
      status: 'Publicada',
      views: 142,
      createdBy: 'Patricia Sánchez'
    },
    {
      id: 6,
      title: 'Energías Renovables para el Hogar',
      description: 'Descubre cómo implementar energías limpias en tu casa',
      date: '25 de Junio, 2026',
      category: 'Ecología',
      status: 'Borrador',
      views: 0,
      createdBy: 'Miguel Torres'
    },
  ];

  const filteredCampaigns = campaigns.filter(campaign =>
    campaign.title.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleCreateCampaign = () => {
    setModalMode('create');
    setSelectedCampaign(null);
    setShowModal(true);
  };

  const handleEditCampaign = (campaign: any) => {
    setModalMode('edit');
    setSelectedCampaign(campaign);
    setShowModal(true);
  };

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
              <h1 className="text-2xl text-white">Gestión de Campañas</h1>
              <p className="text-sm text-white/70">Administra campañas y comunicaciones</p>
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
        {/* Search and Actions */}
        <div className="bg-card rounded-lg border border-border p-6 mb-6">
          <div className="flex flex-col lg:flex-row gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="Buscar campañas por título..."
              />
            </div>
            <button
              onClick={handleCreateCampaign}
              className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity flex items-center gap-2 whitespace-nowrap"
            >
              <Plus className="w-5 h-5" />
              Nueva Campaña
            </button>
          </div>
        </div>

        {/* Campaigns Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {filteredCampaigns.map((campaign) => (
            <div key={campaign.id} className="bg-card rounded-lg border border-border overflow-hidden h-full flex flex-col">
              <div className="aspect-square bg-gradient-to-br from-primary to-[#1a3a1f] flex items-center justify-center flex-shrink-0">
                <span className="text-4xl">📢</span>
              </div>
              <div className="p-4 flex flex-col flex-1">
                <div className="flex items-center justify-between mb-2 gap-2">
                  <span className="inline-flex px-2 py-1 rounded text-xs bg-primary/10 text-primary whitespace-nowrap">
                    {campaign.category}
                  </span>
                  <span className={`inline-flex px-2 py-1 rounded text-xs whitespace-nowrap ${
                    campaign.status === 'Publicada'
                      ? 'bg-[#66bb6a]/10 text-[#2d7a3e]'
                      : 'bg-[#f59e0b]/10 text-[#d97706]'
                  }`}>
                    {campaign.status}
                  </span>
                </div>

                <h3 className="text-sm font-semibold mb-1 line-clamp-2">{campaign.title}</h3>
                <p className="text-xs text-muted-foreground mb-3 line-clamp-2 flex-1">{campaign.description}</p>

                <div className="flex items-center justify-between text-xs text-muted-foreground mb-3 pb-3 border-t border-border">
                  <span className="truncate">{campaign.date}</span>
                  <span className="flex-shrink-0 ml-2">{campaign.views}</span>
                </div>

                <div className="flex gap-1">
                  <button
                    className="flex-1 px-2 py-1 rounded border border-border hover:bg-secondary transition-colors flex items-center justify-center gap-1 text-xs"
                  >
                    <Eye className="w-3 h-3" />
                    Ver
                  </button>
                  <button
                    onClick={() => handleEditCampaign(campaign)}
                    className="flex-1 px-2 py-1 rounded border border-border hover:bg-secondary transition-colors flex items-center justify-center gap-1 text-xs"
                  >
                    <Edit className="w-3 h-3" />
                  </button>
                  <button
                    className="px-2 py-1 rounded border border-destructive/20 text-destructive hover:bg-destructive/10 transition-colors"
                  >
                    <Trash2 className="w-3 h-3" />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>

        {filteredCampaigns.length === 0 && (
          <div className="text-center py-12">
            <p className="text-muted-foreground text-lg">No se encontraron campañas</p>
          </div>
        )}
      </main>

      {/* Create/Edit Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-card border-b border-border p-6 flex items-center justify-between">
              <h2 className="text-2xl">
                {modalMode === 'create' ? 'Nueva Campaña' : 'Editar Campaña'}
              </h2>
              <button
                onClick={() => setShowModal(false)}
                className="p-2 rounded-lg hover:bg-secondary transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6">
              <form className="space-y-4">
                <div>
                  <label className="block mb-2">Título de la Campaña</label>
                  <input
                    type="text"
                    defaultValue={selectedCampaign?.title}
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="Ej: Día Mundial del Medio Ambiente"
                    required
                  />
                </div>

                <div>
                  <label className="block mb-2">Categoría</label>
                  <select
                    defaultValue={selectedCampaign?.category}
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    required
                  >
                    <option value="">Seleccionar categoría</option>
                    <option value="Ambiental">Ambiental</option>
                    <option value="Educación Ambiental">Educación Ambiental</option>
                    <option value="Salud">Salud</option>
                    <option value="Social">Social</option>
                    <option value="Económico">Económico</option>
                  </select>
                </div>

                <div>
                  <label className="block mb-2">Fecha del Evento</label>
                  <input
                    type="date"
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    required
                  />
                </div>

                <div>
                  <label className="block mb-2">Descripción Breve</label>
                  <textarea
                    defaultValue={selectedCampaign?.description}
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary min-h-[80px]"
                    placeholder="Una breve descripción para la lista..."
                    required
                  />
                </div>

                <div>
                  <label className="block mb-2">Contenido Completo</label>
                  <textarea
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary min-h-[150px]"
                    placeholder="Describe la campaña en detalle..."
                    required
                  />
                </div>

                <div>
                  <label className="block mb-2">Imagen de la Campaña</label>
                  <div className="border-2 border-dashed border-border rounded-lg p-8 text-center hover:border-primary transition-colors cursor-pointer">
                    <Upload className="w-8 h-8 text-muted-foreground mx-auto mb-2" />
                    <p className="text-sm text-muted-foreground mb-1">Haz clic o arrastra una imagen aquí</p>
                    <p className="text-xs text-muted-foreground">PNG, JPG hasta 5MB</p>
                  </div>
                </div>

                <div>
                  <label className="block mb-2">Estado</label>
                  <select
                    defaultValue={selectedCampaign?.status || 'Borrador'}
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    required
                  >
                    <option value="Borrador">Borrador (no visible públicamente)</option>
                    <option value="Publicada">Publicada (visible para todos)</option>
                  </select>
                </div>

                <div className="bg-secondary rounded-lg p-4">
                  <p className="text-sm text-muted-foreground">
                    <strong className="text-foreground">Nota:</strong> Las campañas publicadas serán visibles
                    en la sección pública del sitio para usuarios no autenticados.
                  </p>
                </div>

                <div className="flex justify-end gap-3 pt-4">
                  <button
                    type="button"
                    onClick={() => setShowModal(false)}
                    className="px-6 py-3 rounded-lg border border-border hover:bg-secondary transition-colors"
                  >
                    Cancelar
                  </button>
                  <button
                    type="submit"
                    className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
                  >
                    {modalMode === 'create' ? 'Crear Campaña' : 'Guardar Cambios'}
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
