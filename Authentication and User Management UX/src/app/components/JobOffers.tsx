import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Search, MapPin, Clock, Briefcase, Building2, LogOut, FileText, Sparkles, X, ArrowLeft } from 'lucide-react';

interface JobOffersProps {
  onLogout: () => void;
}

export function JobOffers({ onLogout }: JobOffersProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [appliedIds, setAppliedIds] = useState<number[]>([]);
  const [showApplyModal, setShowApplyModal] = useState<any>(null);

  const jobOffers = [
    {
      id: 1,
      title: 'Ayudante de Panadería',
      company: 'Panadería Don José',
      description: 'Buscamos ayudante de panadería con conocimientos básicos de amasado y horneado',
      location: 'Alta Gracia, Córdoba',
      type: 'Tiempo completo',
      salary: 'A convenir',
      requirements: ['Experiencia en panadería (deseable)', 'Disponibilidad horaria', 'Buena predisposición'],
      postedDate: '10 de Mayo, 2026',
      deadline: '30 de Mayo, 2026',
      suggested: true,
    },
    {
      id: 2,
      title: 'Costurero/a',
      company: 'Taller de Costura María',
      description: 'Se solicita persona con conocimientos de costura y confección para taller textil',
      location: 'Alta Gracia, Córdoba',
      type: 'Medio tiempo',
      salary: '$150,000/mes',
      requirements: ['Manejo de máquina de coser', 'Conocimientos de confección', 'Responsabilidad'],
      postedDate: '12 de Mayo, 2026',
      deadline: '25 de Mayo, 2026',
      suggested: false,
    },
    {
      id: 3,
      title: 'Ayudante de Cocina',
      company: 'Restaurante El Buen Sabor',
      description: 'Incorporamos ayudante de cocina para preparación de alimentos y limpieza',
      location: 'Villa Carlos Paz, Córdoba',
      type: 'Tiempo completo',
      salary: '$180,000/mes',
      requirements: ['Conocimientos básicos de cocina', 'Manipulación de alimentos', 'Trabajo en equipo'],
      postedDate: '8 de Mayo, 2026',
      deadline: '22 de Mayo, 2026',
      suggested: true,
    },
  ];

  const filteredOffers = jobOffers.filter(offer =>
    offer.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
    offer.company.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const suggestedOffers = jobOffers.filter(o => o.suggested);

  const handleApply = (offer: any) => {
    setAppliedIds((prev) => [...prev, offer.id]);
    setShowApplyModal(offer);
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="bg-card border-b border-border p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link to="/dashboard" className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center">
              <img src="/logo.png" alt="Buen Vivir" className="w-8 h-8 object-contain" />
            </Link>
            <div>
              <h1 className="text-2xl">Ofertas Laborales</h1>
              <p className="text-sm text-muted-foreground">Buscá oportunidades y postulate</p>
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
      <main className="container mx-auto px-6 py-8">
        {/* CV reminder */}
        <div className="bg-primary/10 border border-primary/20 rounded-lg p-4 mb-6 flex items-center justify-between gap-4 flex-wrap">
          <div className="flex items-center gap-3">
            <FileText className="w-5 h-5 text-primary flex-shrink-0" />
            <p className="text-sm text-muted-foreground">
              Generá tu CV automáticamente para postularte más rápido a estas ofertas.
            </p>
          </div>
          <Link
            to="/generar-cv"
            className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity text-sm whitespace-nowrap"
          >
            Generar mi CV
          </Link>
        </div>

        {/* Suggested offers */}
        {suggestedOffers.length > 0 && (
          <div className="mb-8">
            <h2 className="text-lg font-semibold mb-3 flex items-center gap-2">
              <Sparkles className="w-5 h-5 text-secondary" />
              Sugeridas para vos
            </h2>
            <div className="flex flex-wrap gap-2">
              {suggestedOffers.map((offer) => (
                <span key={offer.id} className="text-xs px-3 py-1.5 rounded-full bg-secondary/10 text-secondary">
                  {offer.title} · {offer.company}
                </span>
              ))}
            </div>
          </div>
        )}

        {/* Search */}
        <div className="bg-card rounded-lg border border-border p-6 mb-8">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
              placeholder="Buscar por título o empresa..."
            />
          </div>
        </div>

        {/* Job Offers List */}
        <div className="space-y-6">
          {filteredOffers.map((offer) => (
            <div key={offer.id} className="bg-card rounded-lg border border-border p-6 hover:border-primary hover:shadow-lg transition-all">
              <div className="flex flex-col lg:flex-row lg:items-start lg:justify-between gap-4 mb-4">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-2">
                    <h2 className="text-2xl">{offer.title}</h2>
                    {offer.suggested && (
                      <span className="inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full bg-secondary/10 text-secondary">
                        <Sparkles className="w-3 h-3" /> Sugerida
                      </span>
                    )}
                  </div>
                  <div className="flex flex-wrap items-center gap-4 text-muted-foreground mb-4">
                    <div className="flex items-center gap-2">
                      <Building2 className="w-4 h-4" />
                      <span>{offer.company}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <MapPin className="w-4 h-4" />
                      <span>{offer.location}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Briefcase className="w-4 h-4" />
                      <span>{offer.type}</span>
                    </div>
                  </div>
                  <p className="text-muted-foreground mb-4">{offer.description}</p>
                </div>
                <div className="lg:text-right">
                  <div className="inline-flex px-4 py-2 rounded-lg bg-primary/10 text-primary mb-2">
                    {offer.salary}
                  </div>
                </div>
              </div>

              <div className="mb-4">
                <h3 className="mb-2">Requisitos:</h3>
                <ul className="list-disc list-inside space-y-1 text-muted-foreground">
                  {offer.requirements.map((req, index) => (
                    <li key={index}>{req}</li>
                  ))}
                </ul>
              </div>

              <div className="flex items-center justify-between pt-4 border-t border-border">
                <div className="flex items-center gap-4 text-sm text-muted-foreground">
                  <span>Publicado: {offer.postedDate}</span>
                  <span className="flex items-center gap-1">
                    <Clock className="w-4 h-4" />
                    Cierra: {offer.deadline}
                  </span>
                </div>
                <button
                  onClick={() => handleApply(offer)}
                  disabled={appliedIds.includes(offer.id)}
                  className="px-6 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {appliedIds.includes(offer.id) ? 'Postulación enviada' : 'Postularme'}
                </button>
              </div>
            </div>
          ))}
        </div>

        {filteredOffers.length === 0 && (
          <div className="text-center py-12">
            <p className="text-muted-foreground text-lg">No se encontraron ofertas laborales</p>
          </div>
        )}
      </main>

      {/* Apply confirmation modal */}
      {showApplyModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-md p-6">
            <div className="flex items-start justify-between mb-4">
              <h2 className="text-xl font-semibold">¡Postulación enviada!</h2>
              <button onClick={() => setShowApplyModal(null)} className="p-1 rounded-lg hover:bg-secondary transition-colors">
                <X className="w-5 h-5" />
              </button>
            </div>
            <p className="text-muted-foreground mb-6">
              Tu postulación a <strong className="text-foreground">{showApplyModal.title}</strong> en {showApplyModal.company} fue enviada correctamente.
            </p>
            <button
              onClick={() => setShowApplyModal(null)}
              className="w-full px-4 py-2.5 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
            >
              Cerrar
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
