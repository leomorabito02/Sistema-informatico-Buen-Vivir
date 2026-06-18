import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Leaf, Search, Calendar, ArrowLeft, Megaphone } from 'lucide-react';
import { Button } from 'primereact/button';

export function PublicCampaigns() {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCampaign, setSelectedCampaign] = useState<any>(null);

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
    <div className="min-h-screen bg-white">
      {/* Header - Greenpeace Style */}
      <header className="bg-gradient-to-r from-[#2d8659] to-[#1a5c3a] text-white border-b-4 border-white/20">
        <div className="container mx-auto px-6 py-6">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-4">
              <Link to="/" className="w-12 h-12 rounded-lg bg-white flex items-center justify-center">
                <img src="/logo.png" alt="Buen Vivir" className="w-8 h-8 object-contain" />
              </Link>
              <div>
                <h1 className="text-4xl font-bold">Campañas</h1>
                <p className="text-white/80">Iniciativas para el buen vivir</p>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <Link
                to="/"
                className="px-6 py-2 rounded-lg bg-white text-[#2d8659] font-bold hover:bg-white/90 transition-colors flex items-center gap-2"
              >
                Apoyá nuestro trabajo
              </Link>
              <Link
                to="/"
                className="px-4 py-2 rounded-lg bg-white/20 border border-white text-white hover:bg-white/30 transition-colors flex items-center gap-2"
              >
                <ArrowLeft className="w-4 h-4" />
                Volver
              </Link>
            </div>
          </div>

          {/* Search Bar */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-white/60" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-3 rounded-lg bg-white/20 border border-white/30 text-white placeholder-white/60 focus:outline-none focus:ring-2 focus:ring-white"
              placeholder="Buscar campañas..."
            />
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="bg-white">
        <div className="container mx-auto px-6 py-12">
          {/* Campaigns Grid - Greenpeace Style */}
          <div className="space-y-16">
            {filteredCampaigns.map((campaign, index) => (
              <div key={campaign.id} className={`grid grid-cols-1 lg:grid-cols-2 gap-8 items-center ${index % 2 === 1 ? 'lg:flex-row-reverse' : ''}`}>
                {/* Image Section */}
                <div className={index % 2 === 1 ? 'lg:order-2' : ''}>
                  <div className="aspect-video bg-gradient-to-br from-[#2d8659]/20 to-[#1a5c3a]/20 rounded-lg overflow-hidden shadow-lg flex items-center justify-center text-9xl">
                    {campaign.image}
                  </div>
                </div>

                {/* Content Section */}
                <div className={index % 2 === 1 ? 'lg:order-1' : ''}>
                  <span className="inline-flex px-4 py-2 rounded-full text-xs font-bold bg-[#2d8659] text-white mb-4">
                    {campaign.category}
                  </span>
                  
                  <h2 className="text-4xl font-bold text-[#1a1a1a] mb-4 leading-tight">
                    {campaign.title}
                  </h2>
                  
                  <p className="text-lg text-gray-700 mb-6">
                    {campaign.description}
                  </p>

                  <div className="mb-6 pt-6 border-t border-gray-300">
                    <p className="text-base text-gray-600 mb-4">
                      {campaign.content}
                    </p>
                  </div>

                  <div className="flex items-center gap-2 text-gray-600 mb-6">
                    <Calendar className="w-5 h-5" />
                    <span className="font-medium">{campaign.date}</span>
                  </div>

                  <button
                    onClick={() => setSelectedCampaign(campaign)}
                    className="px-8 py-3 bg-white border-2 border-[#2d8659] text-[#2d8659] font-bold rounded-lg hover:bg-[#f5f5f5] transition-colors"
                  >
                    Leé más
                  </button>
                </div>
              </div>
            ))}
          </div>

          {filteredCampaigns.length === 0 && (
            <div className="text-center py-12">
              <p className="text-gray-600 text-lg">No se encontraron campañas</p>
            </div>
          )}
        </div>
      </main>

      {/* Modal for Campaign Details */}
      {selectedCampaign && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg w-full max-w-2xl p-8 max-h-[90vh] overflow-y-auto">
            <div className="flex items-start justify-between mb-6">
              <h2 className="text-3xl font-bold text-[#1a1a1a]">{selectedCampaign.title}</h2>
              <button
                onClick={() => setSelectedCampaign(null)}
                className="text-2xl text-gray-600 hover:text-gray-900"
              >
                ×
              </button>
            </div>

            <div className="mb-6">
              <span className="inline-flex px-4 py-2 rounded-full text-xs font-bold bg-[#2d8659] text-white mb-4">
                {selectedCampaign.category}
              </span>
            </div>

            <p className="text-lg text-gray-700 mb-6">
              {selectedCampaign.description}
            </p>

            <div className="mb-6 p-6 bg-gray-50 rounded-lg border border-gray-200">
              <p className="text-gray-600">
                {selectedCampaign.content}
              </p>
            </div>

            <div className="flex items-center gap-2 text-gray-600 mb-8">
              <Calendar className="w-5 h-5" />
              <span className="font-medium">{selectedCampaign.date}</span>
            </div>

            <div className="flex gap-4">
              <button
                onClick={() => setSelectedCampaign(null)}
                className="flex-1 px-6 py-3 bg-gray-300 text-gray-800 font-bold rounded-lg hover:bg-gray-400 transition-colors"
              >
                Cerrar
              </button>
              <Button label="Registrarme" className="flex-1 bg-[#2d8659] text-white font-bold hover:bg-[#1a5c3a] border-none" />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
