import { Link } from 'react-router-dom';
import { useState } from 'react';
import {
  Briefcase, Megaphone, LogIn, BookOpen, Heart, MapPin, Clock, Phone,
  UtensilsCrossed, Hammer, Sprout, Laptop, Scissors, HeartPulse, Briefcase as BriefcaseIcon, Mic,
  Wheat, Leaf as LeafIcon, Salad, X, User, CheckCircle2, Mail
} from 'lucide-react';
import { Button } from 'primereact/button';

interface LibraryMaterial {
  title: string;
  actionLabel: string;
}

export function LandingPage() {
  const [downloadTarget, setDownloadTarget] = useState<LibraryMaterial | null>(null);
  const [contactForm, setContactForm] = useState({ nombre: '', apellido: '', email: '', telefono: '' });
  const [formError, setFormError] = useState('');
  const [downloadReady, setDownloadReady] = useState(false);

  const openDownloadModal = (material: LibraryMaterial) => {
    setDownloadTarget(material);
    setContactForm({ nombre: '', apellido: '', email: '', telefono: '' });
    setFormError('');
    setDownloadReady(false);
  };

  const closeDownloadModal = () => {
    setDownloadTarget(null);
  };

  const handleConfirmDownload = (e: React.FormEvent) => {
    e.preventDefault();
    const { nombre, apellido, email, telefono } = contactForm;
    if (!nombre.trim() || !apellido.trim() || !email.trim() || !telefono.trim()) {
      setFormError('Completá nombre, apellido, email y teléfono para continuar.');
      return;
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) {
      setFormError('Ingresá un email válido.');
      return;
    }
    const phoneDigits = telefono.replace(/[\s()-]/g, '');
    if (!/^\+?\d{6,15}$/.test(phoneDigits)) {
      setFormError('Ingresá un número de teléfono válido.');
      return;
    }
    setFormError('');
    setDownloadReady(true);
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Hero Section */}
      <header className="relative bg-gradient-to-br from-primary to-[#1D1D1B] text-white overflow-hidden">
        <div className="absolute inset-0 opacity-10">
          <div className="absolute top-0 left-0 w-96 h-96 bg-white rounded-full -translate-x-1/2 -translate-y-1/2" />
          <div className="absolute bottom-0 right-0 w-96 h-96 bg-white rounded-full translate-x-1/2 translate-y-1/2" />
        </div>

        <nav className="relative border-b border-white/10 backdrop-blur-sm">
          <div className="container mx-auto px-6 py-4 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-lg bg-white/20 flex items-center justify-center overflow-hidden">
                <img src="/logo.png" alt="Logo Buen Vivir" className="w-10 h-10 object-contain" />
              </div>
              <span className="text-xl">Buen Vivir</span>
            </div>
            <div className="flex items-center gap-3">

              <Link to="/register?role=empresa">
                <Button label="Ofrecer Empleo" icon={<Briefcase className="w-4 h-4 mr-2" />} className="p-button-outlined text-white border-white/50 hover:bg-white/25" />
              </Link>
              <Link to="/login">
                <Button label="Iniciar Sesión" icon={<LogIn className="w-4 h-4 mr-2" />} className="bg-white text-primary border-none hover:bg-white/90" />
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
              Plataforma de acompañamiento integral para personas del Centro Barrial "La otra orilla".
              Transformamos vidas a través de la educación, el trabajo digno y el desarrollo personal.
            </p>
            <div className="flex flex-wrap gap-4">  
              <Link to="/campanas-publico">
                <Button label="Campañas y Concientización" icon={<Megaphone className="w-5 h-5 mr-2" />} className="p-button-outlined border-2 border-white text-white hover:bg-white/20 backdrop-blur-sm text-lg py-3 px-6" />
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
              <div className="text-5xl text-primary mb-2">30</div>
              <div className="text-muted-foreground">Participantes Activos</div>
            </div>
            <div className="text-center">
              <div className="text-5xl text-primary mb-2">12</div>
              <div className="text-muted-foreground">Cursos Disponibles</div>
            </div>
            <div className="text-center">
              <div className="text-5xl text-primary mb-2">+15</div>
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
                <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center mb-4">
                  <UtensilsCrossed className="w-6 h-6 text-primary" />
                </div>
                <h4 className="text-lg font-semibold mb-2">Gastronomía</h4>
                <p className="text-sm text-muted-foreground mb-4">Técnicas culinarias, nutrición y seguridad alimentaria</p>
                <span className="text-xs bg-primary/10 text-primary px-3 py-1 rounded">12 recursos</span>
              </div>

              {/* Oficios */}
              <div className="bg-card rounded-lg border border-border p-6 hover:border-primary hover:shadow-lg transition-all">
                <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center mb-4">
                  <Hammer className="w-6 h-6 text-primary" />
                </div>
                <h4 className="text-lg font-semibold mb-2">Oficios y Construcción</h4>
                <p className="text-sm text-muted-foreground mb-4">Carpintería, electricidad, plomería y más</p>
                <span className="text-xs bg-primary/10 text-primary px-3 py-1 rounded">8 recursos</span>
              </div>

              {/* Ecología */}
              <div className="bg-card rounded-lg border border-border p-6 hover:border-primary hover:shadow-lg transition-all">
                <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center mb-4">
                  <Sprout className="w-6 h-6 text-primary" />
                </div>
                <h4 className="text-lg font-semibold mb-2">Ecología y Sostenibilidad</h4>
                <p className="text-sm text-muted-foreground mb-4">Agricultura orgánica, reciclaje y energías renovables</p>
                <span className="text-xs bg-primary/10 text-primary px-3 py-1 rounded">10 recursos</span>
              </div>

              {/* Tecnología */}
              <div className="bg-card rounded-lg border border-border p-6 hover:border-primary hover:shadow-lg transition-all">
                <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center mb-4">
                  <Laptop className="w-6 h-6 text-primary" />
                </div>
                <h4 className="text-lg font-semibold mb-2">Tecnología Digital</h4>
                <p className="text-sm text-muted-foreground mb-4">Informática, herramientas web y seguridad digital</p>
                <span className="text-xs bg-primary/10 text-primary px-3 py-1 rounded">9 recursos</span>
              </div>

              {/* Arte y Diseño */}
              <div className="bg-card rounded-lg border border-border p-6 hover:border-primary hover:shadow-lg transition-all">
                <div className="w-12 h-12 rounded-full bg-secondary/10 flex items-center justify-center mb-4">
                  <Scissors className="w-6 h-6 text-secondary" />
                </div>
                <h4 className="text-lg font-semibold mb-2">Arte y Diseño</h4>
                <p className="text-sm text-muted-foreground mb-4">Costura, diseño de moda y expresión creativa</p>
                <span className="text-xs bg-secondary/10 text-secondary px-3 py-1 rounded">7 recursos</span>
              </div>

              {/* Bienestar */}
              <div className="bg-card rounded-lg border border-border p-6 hover:border-primary hover:shadow-lg transition-all">
                <div className="w-12 h-12 rounded-full bg-secondary/10 flex items-center justify-center mb-4">
                  <HeartPulse className="w-6 h-6 text-secondary" />
                </div>
                <h4 className="text-lg font-semibold mb-2">Salud y Bienestar</h4>
                <p className="text-sm text-muted-foreground mb-4">Nutrición, ejercicio y desarrollo personal</p>
                <span className="text-xs bg-secondary/10 text-secondary px-3 py-1 rounded">11 recursos</span>
              </div>

              {/* Administración */}
              <div className="bg-card rounded-lg border border-border p-6 hover:border-primary hover:shadow-lg transition-all">
                <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center mb-4">
                  <BriefcaseIcon className="w-6 h-6 text-primary" />
                </div>
                <h4 className="text-lg font-semibold mb-2">Gestión y Negocios</h4>
                <p className="text-sm text-muted-foreground mb-4">Emprendimiento, finanzas y administración</p>
                <span className="text-xs bg-primary/10 text-primary px-3 py-1 rounded">6 recursos</span>
              </div>

              {/* Comunicación */}
              <div className="bg-card rounded-lg border border-border p-6 hover:border-primary hover:shadow-lg transition-all">
                <div className="w-12 h-12 rounded-full bg-secondary/10 flex items-center justify-center mb-4">
                  <Mic className="w-6 h-6 text-secondary" />
                </div>
                <h4 className="text-lg font-semibold mb-2">Comunicación</h4>
                <p className="text-sm text-muted-foreground mb-4">Oralidad, escritura y habilidades interpersonales</p>
                <span className="text-xs bg-secondary/10 text-secondary px-3 py-1 rounded">5 recursos</span>
              </div>
            </div>
          </div>

          {/* Featured Materials */}
          <div className="mb-20">
            <h3 className="text-3xl font-semibold mb-12 text-center">Recursos Públicos</h3>
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
              {/* Material 1 */}
              <div className="bg-card rounded-lg border border-border overflow-hidden hover:border-primary hover:shadow-xl transition-all">
                <div className="h-40 bg-primary/10 flex items-center justify-center">
                  <Wheat className="w-14 h-14 text-primary" />
                </div>
                <div className="p-6">
                  <div className="flex items-center justify-between mb-3">
                    <span className="px-3 py-1 rounded text-xs font-medium bg-primary/10 text-primary">PDF</span>
                    <span className="text-xs text-muted-foreground">5 MB</span>
                  </div>
                  <h4 className="text-lg font-semibold mb-2">Guía Completa de Panadería Artesanal</h4>
                  <p className="text-sm text-muted-foreground mb-4">
                    Manual profesional con técnicas de amasado, fermentación, horneado y presentación de panes artesanales
                  </p>
                  <div className="flex items-center justify-between text-xs text-muted-foreground mb-4">
                    <span className="flex items-center gap-1"><UtensilsCrossed className="w-3.5 h-3.5" /> Gastronomía</span>
                  </div>
                  <Button label="Descargar Material" className="w-full p-button-outlined" onClick={() => openDownloadModal({ title: 'Guía Completa de Panadería Artesanal', actionLabel: 'Descargar Material' })} />
                </div>
              </div>

              {/* Material 2 */}
              <div className="bg-card rounded-lg border border-border overflow-hidden hover:border-primary hover:shadow-xl transition-all">
                <div className="h-40 bg-primary/10 flex items-center justify-center">
                  <LeafIcon className="w-14 h-14 text-primary" />
                </div>
                <div className="p-6">
                  <div className="flex items-center justify-between mb-3">
                    <span className="px-3 py-1 rounded text-xs font-medium bg-primary/10 text-primary">Guía</span>
                    <span className="text-xs text-muted-foreground">8 MB</span>
                  </div>
                  <h4 className="text-lg font-semibold mb-2">Manual de Huerta Orgánica</h4>
                  <p className="text-sm text-muted-foreground mb-4">
                    Cultivo sustentable en espacios reducidos, preparación de compost, riego eficiente y control biológico de plagas
                  </p>
                  <div className="flex items-center justify-between text-xs text-muted-foreground mb-4">
                    <span className="flex items-center gap-1"><Sprout className="w-3.5 h-3.5" /> Ecología</span>
                  </div>
                  <Button label="Descargar Material" className="w-full p-button-outlined" onClick={() => openDownloadModal({ title: 'Manual de Huerta Orgánica', actionLabel: 'Descargar Material' })} />
                </div>
              </div>

              {/* Material 3 */}
              <div className="bg-card rounded-lg border border-border overflow-hidden hover:border-primary hover:shadow-xl transition-all">
                <div className="h-40 bg-primary/10 flex items-center justify-center">
                  <Laptop className="w-14 h-14 text-primary" />
                </div>
                <div className="p-6">
                  <div className="flex items-center justify-between mb-3">
                    <span className="px-3 py-1 rounded text-xs font-medium bg-primary/10 text-primary">Vídeo</span>
                    <span className="text-xs text-muted-foreground">120 min</span>
                  </div>
                  <h4 className="text-lg font-semibold mb-2">Informática Básica para Principiantes</h4>
                  <p className="text-sm text-muted-foreground mb-4">
                    Serie de videos sobre sistema operativo, navegador web, email y herramientas de productividad esenciales
                  </p>
                  <div className="flex items-center justify-between text-xs text-muted-foreground mb-4">
                    <span className="flex items-center gap-1"><Laptop className="w-3.5 h-3.5" /> Tecnología</span>
                  </div>
                  <Button label="Ver Vídeo" className="w-full p-button-outlined" onClick={() => openDownloadModal({ title: 'Informática Básica para Principiantes', actionLabel: 'Ver Vídeo' })} />
                </div>
              </div>

              {/* Material 4 */}
              <div className="bg-card rounded-lg border border-border overflow-hidden hover:border-primary hover:shadow-xl transition-all">
                <div className="h-40 bg-secondary/10 flex items-center justify-center">
                  <Scissors className="w-14 h-14 text-secondary" />
                </div>
                <div className="p-6">
                  <div className="flex items-center justify-between mb-3">
                    <span className="px-3 py-1 rounded text-xs font-medium bg-secondary/10 text-secondary">Guía</span>
                    <span className="text-xs text-muted-foreground">6 MB</span>
                  </div>
                  <h4 className="text-lg font-semibold mb-2">Técnicas de Costura y Diseño</h4>
                  <p className="text-sm text-muted-foreground mb-4">
                    Puntadas básicas, uso de máquina de coser, lectura de patrones y adaptación de tallas para crear prendas
                  </p>
                  <div className="flex items-center justify-between text-xs text-muted-foreground mb-4">
                    <span className="flex items-center gap-1"><Scissors className="w-3.5 h-3.5" /> Arte y Diseño</span>
                  </div>
                  <Button label="Descargar Material" className="w-full p-button-outlined" onClick={() => openDownloadModal({ title: 'Técnicas de Costura y Diseño', actionLabel: 'Descargar Material' })} />
                </div>
              </div>

              {/* Material 5 */}
              <div className="bg-card rounded-lg border border-border overflow-hidden hover:border-primary hover:shadow-xl transition-all">
                <div className="h-40 bg-primary/10 flex items-center justify-center">
                  <BriefcaseIcon className="w-14 h-14 text-primary" />
                </div>
                <div className="p-6">
                  <div className="flex items-center justify-between mb-3">
                    <span className="px-3 py-1 rounded text-xs font-medium bg-primary/10 text-primary">PDF</span>
                    <span className="text-xs text-muted-foreground">4 MB</span>
                  </div>
                  <h4 className="text-lg font-semibold mb-2">Guía de Emprendimiento</h4>
                  <p className="text-sm text-muted-foreground mb-4">
                    Pasos para crear un negocio, plan de negocios, finanzas básicas, marketing y gestión de clientes
                  </p>
                  <div className="flex items-center justify-between text-xs text-muted-foreground mb-4">
                    <span className="flex items-center gap-1"><BriefcaseIcon className="w-3.5 h-3.5" /> Negocios</span>
                  </div>
                  <Button label="Descargar Material" className="w-full p-button-outlined" onClick={() => openDownloadModal({ title: 'Guía de Emprendimiento', actionLabel: 'Descargar Material' })} />
                </div>
              </div>

              {/* Material 6 */}
              <div className="bg-card rounded-lg border border-border overflow-hidden hover:border-primary hover:shadow-xl transition-all">
                <div className="h-40 bg-secondary/10 flex items-center justify-center">
                  <Salad className="w-14 h-14 text-secondary" />
                </div>
                <div className="p-6">
                  <div className="flex items-center justify-between mb-3">
                    <span className="px-3 py-1 rounded text-xs font-medium bg-secondary/10 text-secondary">Vídeo</span>
                    <span className="text-xs text-muted-foreground">90 min</span>
                  </div>
                  <h4 className="text-lg font-semibold mb-2">Nutrición y Alimentación Saludable</h4>
                  <p className="text-sm text-muted-foreground mb-4">
                    Pirámide nutricional, macronutrientes, recetas saludables, presupuesto alimentario y compras inteligentes
                  </p>
                  <div className="flex items-center justify-between text-xs text-muted-foreground mb-4">
                    <span className="flex items-center gap-1"><HeartPulse className="w-3.5 h-3.5" /> Salud</span>
                  </div>
                  <Button label="Ver Vídeo" className="w-full p-button-outlined" onClick={() => openDownloadModal({ title: 'Nutrición y Alimentación Saludable', actionLabel: 'Ver Vídeo' })} />
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
              <Link to="/register?role=voluntario">
                <Button label="Sumate como Voluntario" className="bg-primary text-primary-foreground border-none hover:opacity-90 px-8 py-3" />
              </Link>
              <Link to="/register?role=empresa">
                <Button label="Ofrecer Trabajo o Pasantía" className="p-button-outlined text-primary border-2 border-primary hover:bg-primary/5 px-8 py-3" />
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


      {/* Footer */}
      <footer className="bg-card border-t border-border py-12">
        <div className="container mx-auto px-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div>
              <div className="flex items-center gap-3 mb-4">
                <div className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center">
                  <img src="/logo.png" alt="Buen Vivir" className="w-8 h-8 object-contain" />
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

      {/* Download Contact Modal (HU-25) */}
      {downloadTarget && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-md">
            <div className="border-b border-border p-6 flex items-start justify-between">
              <div>
                <h2 className="text-xl font-semibold mb-1">
                  {downloadReady ? '¡Listo!' : 'Antes de continuar'}
                </h2>
                <p className="text-sm text-muted-foreground">{downloadTarget.title}</p>
              </div>
              <button
                onClick={closeDownloadModal}
                className="p-2 rounded-lg hover:bg-secondary transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6">
              {!downloadReady ? (
                <form onSubmit={handleConfirmDownload} className="space-y-4">
                  <p className="text-sm text-muted-foreground">
                    Dejanos tus datos de contacto para habilitar la descarga de este material gratuito.
                  </p>

                  <div>
                    <label className="block mb-2 text-sm">Nombre</label>
                    <div className="relative">
                      <User className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                      <input
                        type="text"
                        value={contactForm.nombre}
                        onChange={(e) => setContactForm({ ...contactForm, nombre: e.target.value })}
                        className="w-full pl-9 pr-4 py-2.5 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                        placeholder="Tu nombre"
                        required
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block mb-2 text-sm">Apellido</label>
                    <div className="relative">
                      <User className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                      <input
                        type="text"
                        value={contactForm.apellido}
                        onChange={(e) => setContactForm({ ...contactForm, apellido: e.target.value })}
                        className="w-full pl-9 pr-4 py-2.5 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                        placeholder="Tu apellido"
                        required
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block mb-2 text-sm">Email</label>
                    <div className="relative">
                      <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                      <input
                        type="email"
                        value={contactForm.email}
                        onChange={(e) => setContactForm({ ...contactForm, email: e.target.value })}
                        className="w-full pl-9 pr-4 py-2.5 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                        placeholder="tu@email.com"
                        required
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block mb-2 text-sm">Teléfono</label>
                    <div className="relative">
                      <Phone className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                      <input
                        type="tel"
                        value={contactForm.telefono}
                        onChange={(e) => setContactForm({ ...contactForm, telefono: e.target.value })}
                        className="w-full pl-9 pr-4 py-2.5 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                        placeholder="+54 9 3547 XXX-XXX"
                        required
                      />
                    </div>
                  </div>

                  {formError && (
                    <p className="text-sm text-destructive">{formError}</p>
                  )}

                  <div className="flex gap-3 pt-2">
                    <button
                      type="button"
                      onClick={closeDownloadModal}
                      className="flex-1 px-4 py-2.5 rounded-lg border border-border hover:bg-secondary transition-colors text-sm"
                    >
                      Cancelar
                    </button>
                    <button
                      type="submit"
                      className="flex-1 px-4 py-2.5 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity text-sm"
                    >
                      Continuar
                    </button>
                  </div>
                </form>
              ) : (
                <div className="text-center py-4">
                  <div className="w-14 h-14 rounded-full bg-primary/10 flex items-center justify-center mx-auto mb-4">
                    <CheckCircle2 className="w-7 h-7 text-primary" />
                  </div>
                  <p className="text-muted-foreground mb-6">
                    ¡Gracias, {contactForm.nombre}! Ya podés {downloadTarget.actionLabel.toLowerCase()}.
                  </p>
                  <button
                    onClick={closeDownloadModal}
                    className="w-full px-4 py-2.5 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity text-sm"
                  >
                    {downloadTarget.actionLabel}
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
