import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Leaf, Search, Calendar, ArrowLeft, Megaphone } from 'lucide-react';

export function PublicCampaigns() {
  const [searchQuery, setSearchQuery] = useState('');

  const campaigns = [
    {
      id: 1,
      title: 'Día Mundial del Medio Ambiente',
      description: 'Jornada de concientización sobre el cuidado del medio ambiente y la importancia del reciclaje. Invitamos a toda la comunidad a participar de actividades educativas y de plantación de árboles.',
      image: '🌍',
      date: '5 de Junio, 2026',
      category: 'Ambiental',
      content: 'Únete a nosotros en esta jornada especial donde aprenderemos sobre la importancia de cuidar nuestro planeta. Habrá talleres de reciclaje, plantación de árboles nativos y charlas sobre sustentabilidad.'
    },
    {
      id: 2,
      title: 'Taller de Compostaje Comunitario',
      description: 'Aprende a hacer compost en casa y reduce tus residuos orgánicos. Un taller práctico para toda la familia donde enseñaremos técnicas de compostaje doméstico.',
      image: '♻️',
      date: '15 de Mayo, 2026',
      category: 'Educación Ambiental',
      content: 'El compostaje es una forma simple y efectiva de reducir residuos y crear abono natural para tus plantas. En este taller aprenderás qué materiales compostar, cómo mantener tu compostera y cómo usar el compost resultante.'
    },
    {
      id: 3,
      title: 'Campaña de Concientización sobre Alimentación Saludable',
      description: 'Serie de charlas y talleres sobre nutrición, hábitos alimenticios saludables y cocina económica nutritiva.',
      image: '🥗',
      date: '20 de Mayo, 2026',
      category: 'Salud',
      content: 'Una alimentación saludable es fundamental para el bienestar. Aprende a planificar comidas nutritivas y económicas, conoce los grupos alimenticios y descubre recetas saludables fáciles de preparar.'
    },
    {
      id: 4,
      title: 'Inclusión Social y Derechos Humanos',
      description: 'Jornada de reflexión y diálogo sobre inclusión social, equidad y derechos humanos en nuestra comunidad.',
      image: '🤝',
      date: '18 de Mayo, 2026',
      category: 'Social',
      content: 'Un espacio de encuentro para reflexionar sobre la importancia de la inclusión social, el respeto a la diversidad y la construcción de una sociedad más justa y equitativa para todos.'
    },
  ];

  const filteredCampaigns = campaigns.filter(campaign =>
    campaign.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
    campaign.description.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="bg-card border-b border-border">
        <div className="container mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link to="/" className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center">
                <Leaf className="w-6 h-6 text-primary-foreground" />
              </Link>
              <div>
                <h1 className="text-2xl">Campañas y Concientización</h1>
                <p className="text-sm text-muted-foreground">Iniciativas ambientales y sociales</p>
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
              placeholder="Buscar campañas..."
            />
          </div>
        </div>

        {/* Campaigns Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {filteredCampaigns.map((campaign) => (
            <div key={campaign.id} className="bg-card rounded-lg border border-border overflow-hidden hover:border-primary hover:shadow-lg transition-all">
              <div className="aspect-video bg-gradient-to-br from-primary to-[#1a3a1f] flex items-center justify-center text-9xl">
                {campaign.image}
              </div>
              <div className="p-6">
                <div className="flex items-center justify-between mb-3">
                  <span className="inline-flex px-3 py-1 rounded-full text-xs bg-primary/10 text-primary">
                    {campaign.category}
                  </span>
                  <span className="flex items-center gap-1 text-sm text-muted-foreground">
                    <Calendar className="w-4 h-4" />
                    {campaign.date}
                  </span>
                </div>
                <h2 className="text-2xl mb-3">{campaign.title}</h2>
                <p className="text-muted-foreground mb-4">{campaign.description}</p>
                <div className="pt-4 border-t border-border">
                  <p className="text-sm">{campaign.content}</p>
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
    </div>
  );
}
