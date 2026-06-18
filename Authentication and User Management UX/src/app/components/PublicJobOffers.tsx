import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Leaf, Search, MapPin, Clock, Briefcase, Building2, ArrowLeft } from 'lucide-react';

export function PublicJobOffers() {
  const [searchQuery, setSearchQuery] = useState('');

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
      deadline: '30 de Mayo, 2026'
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
      deadline: '25 de Mayo, 2026'
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
      deadline: '22 de Mayo, 2026'
    },
  ];

  const filteredOffers = jobOffers.filter(offer =>
    offer.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
    offer.company.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="bg-card border-b border-border">
        <div className="container mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link to="/" className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center">
                <img src="/logo.png" alt="Buen Vivir" className="w-8 h-8 object-contain" />
              </Link>
              <div>
                <h1 className="text-2xl">Ofertas Laborales</h1>
                <p className="text-sm text-muted-foreground">Encuentra tu próxima oportunidad</p>
              </div>
            </div>
            <Link
              to="/"
              className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors flex items-center gap-2"
            >
              <ArrowLeft className="w-4 h-4" />
              Volver al Inicio
            </Link>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-6 py-12">
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
                  <h2 className="text-2xl mb-2">{offer.title}</h2>
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
                <Link
                  to="/login"
                  className="px-6 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
                >
                  Postularme
                </Link>
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
    </div>
  );
}
