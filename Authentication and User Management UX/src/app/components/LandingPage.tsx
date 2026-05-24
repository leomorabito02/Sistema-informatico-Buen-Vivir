import { Link } from 'react-router-dom';
import { Leaf, Briefcase, Megaphone, LogIn, Users, BookOpen, Heart, TrendingUp, Award, MapPin, Clock, Phone } from 'lucide-react';

const logoImg = new URL('./logo.png', import.meta.url).href;

export function LandingPage() {
  return (
    <div className="min-h-screen bg-background">
      {/* Hero Section */}
      <header className="relative bg-gradient-to-br from-primary to-[#1a3a1f] text-white overflow-hidden">
        <div className="absolute inset-0 opacity-10">
          <div className="absolute top-0 left-0 w-96 h-96 bg-white rounded-full -translate-x-1/2 -translate-y-1/2" />
          <div className="absolute bottom-0 right-0 w-96 h-96 bg-white rounded-full translate-x-1/2 translate-y-1/2" />
        </div>

        <nav className="relative border-b border-white/10 backdrop-blur-sm">
          <div className="container mx-auto px-6 py-4 flex items-center justify-between">
            <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-white/20 flex items-center justify-center overflow-hidden">
  <img src={logoImg} alt="Logo Buen Vivir" className="w-8 h-8 object-contain" />
</div>
              <span className="text-xl">Buen Vivir</span>
            </div>
            <div className="flex items-center gap-3">
              <a
                href="#donar"
                className="px-5 py-2 bg-white/15 border border-white/50 text-white rounded-lg hover:bg-white/25 transition-colors flex items-center gap-2"
              >
                <Heart className="w-4 h-4" />
                Donar
              </a>
              <Link
                to="/register?role=empresa"
                className="px-5 py-2 bg-white/15 border border-white/50 text-white rounded-lg hover:bg-white/25 transition-colors flex items-center gap-2"
              >
                <Briefcase className="w-4 h-4" />
                Ofrecer Empleo
              </Link>
              <Link
                to="/login"
                className="px-6 py-2 bg-white text-primary rounded-lg hover:bg-white/90 transition-colors flex items-center gap-2"
              >
                <LogIn className="w-4 h-4" />
                Iniciar Sesión
              </Link>
            </div>
          </div>
        </nav>

        <div className="relative container mx-auto px-6 py-24">
          <div className="max-w-3xl">
            <h1 className="text-5xl md:text-6xl mb-6">
              Centro de Ecología Integral
            </h1>
            <p className="text-xl md:text-2xl mb-4 text-white/90">
              La Otra Orilla - Alta Gracia
            </p>
            <p className="text-lg mb-8 text-white/80">
              Plataforma de acompañamiento integral para personas en situación de vulnerabilidad.
              Transformamos vidas a través de la educación, el trabajo digno y el desarrollo personal.
            </p>
            <div className="flex flex-wrap gap-4">  
              <Link
                to="/campanas-publico"
                className="px-8 py-4 bg-white/10 border-2 border-white rounded-lg hover:bg-white/20 transition-colors flex items-center gap-2 text-lg backdrop-blur-sm"
              >
                <Megaphone className="w-5 h-5" />
                Campañas y Concientización
              </Link>
            </div>
          </div>
        </div>
      </header>

      {/* Pillars Section */}
      <section className="py-20 bg-card">
        <div className="container mx-auto px-6">
          <div className="text-center mb-16">
            <h2 className="text-4xl mb-4">Nuestros Pilares de Trabajo</h2>
            <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
              Un enfoque integral para el desarrollo humano y la transformación social
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
            <div className="bg-background rounded-lg border border-border p-8 text-center hover:border-primary hover:shadow-lg transition-all">
              <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center mx-auto mb-4">
                <BookOpen className="w-8 h-8 text-primary" />
              </div>
              <h3 className="text-xl mb-3">Educación Integral</h3>
              <p className="text-muted-foreground">
                Cursos y talleres formativos para el desarrollo de habilidades técnicas y personales
              </p>
            </div>

            <div className="bg-background rounded-lg border border-border p-8 text-center hover:border-primary hover:shadow-lg transition-all">
              <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center mx-auto mb-4">
                <Briefcase className="w-8 h-8 text-primary" />
              </div>
              <h3 className="text-xl mb-3">Inserción Laboral</h3>
              <p className="text-muted-foreground">
                Vinculación con oportunidades laborales y acompañamiento en el proceso
              </p>
            </div>

            <div className="bg-background rounded-lg border border-border p-8 text-center hover:border-primary hover:shadow-lg transition-all">
              <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center mx-auto mb-4">
                <Heart className="w-8 h-8 text-primary" />
              </div>
              <h3 className="text-xl mb-3">Promoción Humana</h3>
              <p className="text-muted-foreground">
                Acompañamiento personalizado en el proceso de desarrollo integral de cada persona
              </p>
            </div>

            <div className="bg-background rounded-lg border border-border p-8 text-center hover:border-primary hover:shadow-lg transition-all">
              <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center mx-auto mb-4">
                <Megaphone className="w-8 h-8 text-primary" />
              </div>
              <h3 className="text-xl mb-3">Concientización</h3>
              <p className="text-muted-foreground">
                Campañas de sensibilización ambiental y social para toda la comunidad
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Impact Section */}
      <section className="py-20">
        <div className="container mx-auto px-6">
          <div className="text-center mb-16">
            <h2 className="text-4xl mb-4">Nuestro Impacto</h2>
            <p className="text-xl text-muted-foreground">
              Transformando vidas a través del acompañamiento integral
            </p>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-3 gap-8">
            <div className="text-center">
              <div className="text-5xl text-primary mb-2">248</div>
              <div className="text-muted-foreground">Participantes Activos</div>
            </div>
            <div className="text-center">
              <div className="text-5xl text-primary mb-2">12</div>
              <div className="text-muted-foreground">Cursos Disponibles</div>
            </div>
            <div className="text-center">
              <div className="text-5xl text-primary mb-2">32</div>
              <div className="text-muted-foreground">Voluntarios</div>
            </div>
          </div>
        </div>
      </section>

      {/* How It Works */}
      <section className="py-20 bg-card">
        <div className="container mx-auto px-6">
          <div className="text-center mb-16">
            <h2 className="text-4xl mb-4">¿Cómo Funciona?</h2>
            <p className="text-xl text-muted-foreground">
              Un proceso integral de acompañamiento y desarrollo
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div className="relative">
              <div className="bg-background rounded-lg border border-border p-8">
                <div className="w-12 h-12 rounded-full bg-primary text-primary-foreground flex items-center justify-center mb-4 text-xl">
                  1
                </div>
                <h3 className="text-xl mb-3">Registro y Evaluación</h3>
                <p className="text-muted-foreground">
                  Te damos la bienvenida y evaluamos tus necesidades, intereses y habilidades actuales
                </p>
              </div>
              <div className="hidden md:block absolute top-1/2 -right-4 w-8 h-0.5 bg-border" />
            </div>

            <div className="relative">
              <div className="bg-background rounded-lg border border-border p-8">
                <div className="w-12 h-12 rounded-full bg-primary text-primary-foreground flex items-center justify-center mb-4 text-xl">
                  2
                </div>
                <h3 className="text-xl mb-3">Formación y Desarrollo</h3>
                <p className="text-muted-foreground">
                  Participás en cursos y talleres adaptados a tu perfil, con seguimiento personalizado
                </p>
              </div>
              <div className="hidden md:block absolute top-1/2 -right-4 w-8 h-0.5 bg-border" />
            </div>

            <div className="bg-background rounded-lg border border-border p-8">
              <div className="w-12 h-12 rounded-full bg-primary text-primary-foreground flex items-center justify-center mb-4 text-xl">
                3
              </div>
              <h3 className="text-xl mb-3">Inserción Laboral</h3>
              <p className="text-muted-foreground">
                Te vinculamos con oportunidades laborales acordes a tus habilidades y aspiraciones
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Educational Materials Section - Public Access */}
      <section className="py-24 bg-gradient-to-b from-secondary to-background">
        <div className="container mx-auto px-6">
          <div className="text-center mb-20">
            <h2 className="text-5xl mb-4 font-bold">Biblioteca de Recursos Comunitarios</h2>
            <p className="text-xl text-muted-foreground max-w-3xl mx-auto">
              Conocimientos compartidos por nuestra comunidad para el buen vivir. 
              Materiales gratuitos creados con el propósito de fortalecer capacidades y saberes colectivos.
            </p>
          </div>

          {/* Categories Grid */}
          <div className="mb-20">
            <h3 className="text-3xl font-semibold mb-12 text-center">Categorías de Aprendizaje</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              {/* Gastronomía */}
              <div className="bg-card rounded-lg border border-border p-6 hover:border-primary hover:shadow-lg transition-all">
                <div className="w-12 h-12 rounded-full bg-orange-100 flex items-center justify-center mb-4">
                  <span className="text-2xl">🍳</span>
                </div>
                <h4 className="text-lg font-semibold mb-2">Gastronomía</h4>
                <p className="text-sm text-muted-foreground mb-4">Técnicas culinarias, nutrición y seguridad alimentaria</p>
                <span className="text-xs bg-orange-100 text-orange-700 px-3 py-1 rounded">12 recursos</span>
              </div>

              {/* Oficios */}
              <div className="bg-card rounded-lg border border-border p-6 hover:border-primary hover:shadow-lg transition-all">
                <div className="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center mb-4">
                  <span className="text-2xl">🔨</span>
                </div>
                <h4 className="text-lg font-semibold mb-2">Oficios y Construcción</h4>
                <p className="text-sm text-muted-foreground mb-4">Carpintería, electricidad, plomería y más</p>
                <span className="text-xs bg-blue-100 text-blue-700 px-3 py-1 rounded">8 recursos</span>
              </div>

              {/* Ecología */}
              <div className="bg-card rounded-lg border border-border p-6 hover:border-primary hover:shadow-lg transition-all">
                <div className="w-12 h-12 rounded-full bg-green-100 flex items-center justify-center mb-4">
                  <span className="text-2xl">🌱</span>
                </div>
                <h4 className="text-lg font-semibold mb-2">Ecología y Sostenibilidad</h4>
                <p className="text-sm text-muted-foreground mb-4">Agricultura orgánica, reciclaje y energías renovables</p>
                <span className="text-xs bg-green-100 text-green-700 px-3 py-1 rounded">10 recursos</span>
              </div>

              {/* Tecnología */}
              <div className="bg-card rounded-lg border border-border p-6 hover:border-primary hover:shadow-lg transition-all">
                <div className="w-12 h-12 rounded-full bg-purple-100 flex items-center justify-center mb-4">
                  <span className="text-2xl">💻</span>
                </div>
                <h4 className="text-lg font-semibold mb-2">Tecnología Digital</h4>
                <p className="text-sm text-muted-foreground mb-4">Informática, herramientas web y seguridad digital</p>
                <span className="text-xs bg-purple-100 text-purple-700 px-3 py-1 rounded">9 recursos</span>
              </div>

              {/* Arte y Diseño */}
              <div className="bg-card rounded-lg border border-border p-6 hover:border-primary hover:shadow-lg transition-all">
                <div className="w-12 h-12 rounded-full bg-pink-100 flex items-center justify-center mb-4">
                  <span className="text-2xl">✂️</span>
                </div>
                <h4 className="text-lg font-semibold mb-2">Arte y Diseño</h4>
                <p className="text-sm text-muted-foreground mb-4">Costura, diseño de moda y expresión creativa</p>
                <span className="text-xs bg-pink-100 text-pink-700 px-3 py-1 rounded">7 recursos</span>
              </div>

              {/* Bienestar */}
              <div className="bg-card rounded-lg border border-border p-6 hover:border-primary hover:shadow-lg transition-all">
                <div className="w-12 h-12 rounded-full bg-red-100 flex items-center justify-center mb-4">
                  <span className="text-2xl">❤️</span>
                </div>
                <h4 className="text-lg font-semibold mb-2">Salud y Bienestar</h4>
                <p className="text-sm text-muted-foreground mb-4">Nutrición, ejercicio y desarrollo personal</p>
                <span className="text-xs bg-red-100 text-red-700 px-3 py-1 rounded">11 recursos</span>
              </div>

              {/* Administración */}
              <div className="bg-card rounded-lg border border-border p-6 hover:border-primary hover:shadow-lg transition-all">
                <div className="w-12 h-12 rounded-full bg-indigo-100 flex items-center justify-center mb-4">
                  <span className="text-2xl">📊</span>
                </div>
                <h4 className="text-lg font-semibold mb-2">Gestión y Negocios</h4>
                <p className="text-sm text-muted-foreground mb-4">Emprendimiento, finanzas y administración</p>
                <span className="text-xs bg-indigo-100 text-indigo-700 px-3 py-1 rounded">6 recursos</span>
              </div>

              {/* Comunicación */}
              <div className="bg-card rounded-lg border border-border p-6 hover:border-primary hover:shadow-lg transition-all">
                <div className="w-12 h-12 rounded-full bg-cyan-100 flex items-center justify-center mb-4">
                  <span className="text-2xl">🎤</span>
                </div>
                <h4 className="text-lg font-semibold mb-2">Comunicación</h4>
                <p className="text-sm text-muted-foreground mb-4">Oralidad, escritura y habilidades interpersonales</p>
                <span className="text-xs bg-cyan-100 text-cyan-700 px-3 py-1 rounded">5 recursos</span>
              </div>
            </div>
          </div>

          {/* Featured Materials */}
          <div className="mb-20">
            <h3 className="text-3xl font-semibold mb-12 text-center">Recursos Destacados</h3>
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
              {/* Material 1 */}
              <div className="bg-card rounded-lg border border-border overflow-hidden hover:border-primary hover:shadow-xl transition-all">
                <div className="h-40 bg-gradient-to-br from-orange-100 to-yellow-100 flex items-center justify-center">
                  <span className="text-6xl">🍞</span>
                </div>
                <div className="p-6">
                  <div className="flex items-center justify-between mb-3">
                    <span className="px-3 py-1 rounded text-xs font-medium bg-orange-100 text-orange-700">PDF</span>
                    <span className="text-xs text-muted-foreground">5 MB</span>
                  </div>
                  <h4 className="text-lg font-semibold mb-2">Guía Completa de Panadería Artesanal</h4>
                  <p className="text-sm text-muted-foreground mb-4">
                    Manual profesional con técnicas de amasado, fermentación, horneado y presentación de panes artesanales
                  </p>
                  <div className="flex items-center justify-between text-xs text-muted-foreground mb-4">
                    <span>📚 Gastronomía</span>
                    
                  </div>
                  <button className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors w-full text-sm font-medium">
                    Descargar Material
                  </button>
                </div>
              </div>

              {/* Material 2 */}
              <div className="bg-card rounded-lg border border-border overflow-hidden hover:border-primary hover:shadow-xl transition-all">
                <div className="h-40 bg-gradient-to-br from-green-100 to-lime-100 flex items-center justify-center">
                  <span className="text-6xl">🌿</span>
                </div>
                <div className="p-6">
                  <div className="flex items-center justify-between mb-3">
                    <span className="px-3 py-1 rounded text-xs font-medium bg-green-100 text-green-700">Guía</span>
                    <span className="text-xs text-muted-foreground">8 MB</span>
                  </div>
                  <h4 className="text-lg font-semibold mb-2">Manual de Huerta Orgánica</h4>
                  <p className="text-sm text-muted-foreground mb-4">
                    Cultivo sustentable en espacios reducidos, preparación de compost, riego eficiente y control biológico de plagas
                  </p>
                  <div className="flex items-center justify-between text-xs text-muted-foreground mb-4">
                    <span>🌱 Ecología</span>
                    
                  </div>
                  <button className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors w-full text-sm font-medium">
                    Descargar Material
                  </button>
                </div>
              </div>

              {/* Material 3 */}
              <div className="bg-card rounded-lg border border-border overflow-hidden hover:border-primary hover:shadow-xl transition-all">
                <div className="h-40 bg-gradient-to-br from-blue-100 to-cyan-100 flex items-center justify-center">
                  <span className="text-6xl">💻</span>
                </div>
                <div className="p-6">
                  <div className="flex items-center justify-between mb-3">
                    <span className="px-3 py-1 rounded text-xs font-medium bg-blue-100 text-blue-700">Vídeo</span>
                    <span className="text-xs text-muted-foreground">120 min</span>
                  </div>
                  <h4 className="text-lg font-semibold mb-2">Informática Básica para Principiantes</h4>
                  <p className="text-sm text-muted-foreground mb-4">
                    Serie de videos sobre sistema operativo, navegador web, email y herramientas de productividad esenciales
                  </p>
                  <div className="flex items-center justify-between text-xs text-muted-foreground mb-4">
                    <span>💻 Tecnología</span>
                    
                  </div>
                  <button className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors w-full text-sm font-medium">
                    Ver Vídeo
                  </button>
                </div>
              </div>

              {/* Material 4 */}
              <div className="bg-card rounded-lg border border-border overflow-hidden hover:border-primary hover:shadow-xl transition-all">
                <div className="h-40 bg-gradient-to-br from-pink-100 to-rose-100 flex items-center justify-center">
                  <span className="text-6xl">✂️</span>
                </div>
                <div className="p-6">
                  <div className="flex items-center justify-between mb-3">
                    <span className="px-3 py-1 rounded text-xs font-medium bg-pink-100 text-pink-700">Guía</span>
                    <span className="text-xs text-muted-foreground">6 MB</span>
                  </div>
                  <h4 className="text-lg font-semibold mb-2">Técnicas de Costura y Diseño</h4>
                  <p className="text-sm text-muted-foreground mb-4">
                    Puntadas básicas, uso de máquina de coser, lectura de patrones y adaptación de tallas para crear prendas
                  </p>
                  <div className="flex items-center justify-between text-xs text-muted-foreground mb-4">
                    <span>👗 Arte y Diseño</span>
                    
                  </div>
                  <button className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors w-full text-sm font-medium">
                    Descargar Material
                  </button>
                </div>
              </div>

              {/* Material 5 */}
              <div className="bg-card rounded-lg border border-border overflow-hidden hover:border-primary hover:shadow-xl transition-all">
                <div className="h-40 bg-gradient-to-br from-purple-100 to-indigo-100 flex items-center justify-center">
                  <span className="text-6xl">📊</span>
                </div>
                <div className="p-6">
                  <div className="flex items-center justify-between mb-3">
                    <span className="px-3 py-1 rounded text-xs font-medium bg-purple-100 text-purple-700">PDF</span>
                    <span className="text-xs text-muted-foreground">4 MB</span>
                  </div>
                  <h4 className="text-lg font-semibold mb-2">Guía de Emprendimiento</h4>
                  <p className="text-sm text-muted-foreground mb-4">
                    Pasos para crear un negocio, plan de negocios, finanzas básicas, marketing y gestión de clientes
                  </p>
                  <div className="flex items-center justify-between text-xs text-muted-foreground mb-4">
                    <span>💼 Negocios</span>
                   
                  </div>
                  <button className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors w-full text-sm font-medium">
                    Descargar Material
                  </button>
                </div>
              </div>

              {/* Material 6 */}
              <div className="bg-card rounded-lg border border-border overflow-hidden hover:border-primary hover:shadow-xl transition-all">
                <div className="h-40 bg-gradient-to-br from-red-100 to-orange-100 flex items-center justify-center">
                  <span className="text-6xl">❤️</span>
                </div>
                <div className="p-6">
                  <div className="flex items-center justify-between mb-3">
                    <span className="px-3 py-1 rounded text-xs font-medium bg-red-100 text-red-700">Vídeo</span>
                    <span className="text-xs text-muted-foreground">90 min</span>
                  </div>
                  <h4 className="text-lg font-semibold mb-2">Nutrición y Alimentación Saludable</h4>
                  <p className="text-sm text-muted-foreground mb-4">
                    Pirámide nutricional, macronutrientes, recetas saludables, presupuesto alimentario y compras inteligentes
                  </p>
                  <div className="flex items-center justify-between text-xs text-muted-foreground mb-4">
                    <span>🥗 Salud</span>
                   
                  </div>
                  <button className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors w-full text-sm font-medium">
                    Ver Vídeo
                  </button>
                </div>
              </div>
            </div>
          </div>

          {/* CTA Section */}
          <div className="bg-gradient-to-r from-primary/10 to-primary/5 rounded-2xl border border-primary/20 p-12 text-center">
            <h3 className="text-2xl font-semibold mb-4">Sé Parte del Cambio</h3>
            <p className="text-muted-foreground mb-8 max-w-3xl mx-auto">
              Estamos buscando voluntarios comprometidos con acompañar procesos de reinserción social y empresas que crean en el potencial de las personas. 
              Si querés ofrecer trabajo, pasantías o voluntariado para ayudar a transformar vidas, ¡te necesitamos!
            </p>
            <div className="flex flex-wrap gap-4 justify-center">
              <Link
                to="/register?role=voluntario"
                className="px-8 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity font-medium"
              >
                Sumate como Voluntario
              </Link>
              <Link
                to="/register?role=empresa"
                className="px-8 py-3 border-2 border-primary text-primary rounded-lg hover:bg-primary/5 transition-colors font-medium"
              >
                Ofrecer Trabajo o Pasantía
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* ── NUEVA: Ubicación con Google Maps ── */}
      <section className="py-20">
        <div className="container mx-auto px-6">
          <div className="text-center mb-12">
            <h2 className="text-4xl mb-4">¿Dónde estamos?</h2>
            <p className="text-xl text-muted-foreground">Encontranos en Alta Gracia, Córdoba</p>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 items-start">
            {/* Mapa embed — reemplazá el src con el embed URL de tu dirección exacta */}
            <div className="lg:col-span-2 rounded-lg overflow-hidden border border-border">
              <iframe
                title="Ubicación Centro Buen Vivir - Alta Gracia"
                src="https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d26730.2!2d-64.4295!3d-31.6554!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x943285b2e7d3f2e5%3A0xb0a1f3e0c2d4e5f6!2sAlta%20Gracia%2C%20C%C3%B3rdoba!5e0!3m2!1ses-419!2sar!4v1680000000000!5m2!1ses-419!2sar"
                width="100%"
                height="420"
                style={{ border: 0, display: 'block' }}
                allowFullScreen
                loading="lazy"
                referrerPolicy="no-referrer-when-downgrade"
              />
            </div>

            {/* Info de contacto */}
            <div className="space-y-6 pt-2">
              <div className="flex items-start gap-4">
                <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center flex-shrink-0">
                  <MapPin className="w-5 h-5 text-primary" />
                </div>
                <div>
                  <p className="mb-1">Dirección</p>
                  <p className="text-sm text-muted-foreground">
                    La Otra Orilla<br />Alta Gracia, Córdoba, Argentina
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-4">
                <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center flex-shrink-0">
                  <Clock className="w-5 h-5 text-primary" />
                </div>
                <div>
                  <p className="mb-1">Horario de atención</p>
                  <p className="text-sm text-muted-foreground">
                    Lunes a Viernes<br />9:00 – 17:00 hs
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-4">
                <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center flex-shrink-0">
                  <Phone className="w-5 h-5 text-primary" />
                </div>
                <div>
                  <p className="mb-1">Contacto</p>
                  <p className="text-sm text-muted-foreground">
                    +54 9 3547 XXX-XXX<br />contacto@buenvivir.org
                  </p>
                </div>
              </div>

              <a
                href="https://maps.google.com/?q=Alta+Gracia,Cordoba,Argentina"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-2 px-5 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity text-sm"
              >
                <MapPin className="w-4 h-4" />
                Abrir en Google Maps
              </a>
            </div>
          </div>
        </div>
      </section>

      {/* ── NUEVA: Sección de Donación ── */}
      <section id="donar" className="py-20 bg-card">
        <div className="container mx-auto px-6">
          <div className="max-w-2xl mx-auto text-center">
            <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center mx-auto mb-6">
              <Heart className="w-8 h-8 text-primary" />
            </div>
            <h2 className="text-4xl mb-4">Apoyá nuestro trabajo</h2>
            <p className="text-xl text-muted-foreground mb-10">
              Tu donación nos permite seguir acompañando a personas en situación de vulnerabilidad.
              Cada aporte, por pequeño que sea, transforma vidas.
            </p>

            <div className="flex flex-wrap gap-3 justify-center mb-6">
              {['$500', '$1.000', '$2.000', 'Otro monto'].map((monto) => (
                <button
                  key={monto}
                  className="px-6 py-3 rounded-lg border border-border hover:border-primary hover:bg-primary/5 transition-colors"
                >
                  {monto}
                </button>
              ))}
            </div>

            <div className="flex flex-wrap gap-4 justify-center">
              <button className="px-12 py-4 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity text-lg flex items-center gap-3">
                <Heart className="w-5 h-5" />
                Donar ahora
              </button>
            </div>

            <p className="text-sm text-muted-foreground mt-4">
              Procesamos donaciones de forma segura. Podés cancelar cuando quieras.
            </p>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-card border-t border-border py-12">
        <div className="container mx-auto px-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div>
              <div className="flex items-center gap-3 mb-4">
                <div className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center">
                  <Leaf className="w-6 h-6 text-primary-foreground" />
                </div>
                <span className="text-xl">Buen Vivir</span>
              </div>
              <p className="text-muted-foreground">
                Centro de Ecología Integral<br />
                La Otra Orilla - Alta Gracia
              </p>
            </div>

            <div>
              <h3 className="mb-4">Enlaces</h3>
              <div className="space-y-2">
                <Link to="/ofertas-laborales-publico" className="block text-muted-foreground hover:text-primary">
                  Ofertas Laborales
                </Link>
                <Link to="/campanas-publico" className="block text-muted-foreground hover:text-primary">
                  Campañas
                </Link>
                <a href="#donar" className="block text-muted-foreground hover:text-primary">
                  Donar
                </a>
                <Link to="/login" className="block text-muted-foreground hover:text-primary">
                  Iniciar Sesión
                </Link>
              </div>
            </div>

            <div>
              <h3 className="mb-4">Contacto</h3>
              <p className="text-muted-foreground">
                Email: contacto@buenvivir.org<br />
                Teléfono: +54 9 3547 XXX-XXX<br />
                Alta Gracia, Córdoba
              </p>
            </div>
          </div>

          <div className="mt-8 pt-8 border-t border-border text-center text-muted-foreground">
            <p>© 2026 Buen Vivir - Centro de Ecología Integral. Todos los derechos reservados.</p>
          </div>
        </div>
      </footer>
    </div>
  );
}
