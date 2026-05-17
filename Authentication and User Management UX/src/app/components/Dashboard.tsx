import { Link } from 'react-router-dom';
import { Leaf, BookOpen, Users, BarChart3, LogOut, Menu, Bell, Settings, GraduationCap, Award, TrendingUp, Megaphone } from 'lucide-react';
import { useState } from 'react';

interface DashboardProps {
  userRole: 'admin' | 'participante' | 'voluntario' | 'empresa' | 'docente';
  onLogout: () => void;
}

export function Dashboard({ userRole, onLogout }: DashboardProps) {
  const [sidebarOpen, setSidebarOpen] = useState(true);

  const stats = {
    participante: [
      { label: 'Cursos Inscritos', value: '5', icon: BookOpen, color: 'bg-primary' },
      { label: 'Progreso Promedio', value: '68%', icon: TrendingUp, color: 'bg-[#66bb6a]' },
      { label: 'Certificados', value: '2', icon: Award, color: 'bg-[#81c784]' },
      { label: 'Horas de Estudio', value: '127', icon: GraduationCap, color: 'bg-[#a5d6a7]' },
    ],
    admin: [
      { label: 'Total Participantes', value: '248', icon: Users, color: 'bg-primary' },
      { label: 'Cursos Activos', value: '12', icon: BookOpen, color: 'bg-[#66bb6a]' },
      { label: 'Voluntarios', value: '32', icon: Users, color: 'bg-[#81c784]' },
      { label: 'Tasa de Asistencia', value: '85%', icon: BarChart3, color: 'bg-[#a5d6a7]' },
    ],
    voluntario: [
      { label: 'Participantes Asignados', value: '15', icon: Users, color: 'bg-primary' },
      { label: 'Sesiones Esta Semana', value: '8', icon: BookOpen, color: 'bg-[#66bb6a]' },
      { label: 'Tarea Pendientes', value: '3', icon: BarChart3, color: 'bg-[#81c784]' },
    ],
    docente: [
      { label: 'Mis Cursos', value: '3', icon: BookOpen, color: 'bg-primary' },
      { label: 'Total Estudiantes', value: '45', icon: Users, color: 'bg-[#66bb6a]' },
      { label: 'Materiales Subidos', value: '28', icon: GraduationCap, color: 'bg-[#81c784]' },
    ],
    empresa: [
      { label: 'Ofertas Activas', value: '7', icon: BookOpen, color: 'bg-primary' },
      { label: 'Candidatos', value: '23', icon: Users, color: 'bg-[#66bb6a]' },
      { label: 'Contrataciones', value: '5', icon: Award, color: 'bg-[#81c784]' },
    ],
  };

  const currentStats = stats[userRole] || stats.participante;

  const menuItems = {
    participante: [
      { label: 'Dashboard', path: '/dashboard', icon: BarChart3 },
      { label: 'Catálogo de Cursos', path: '/courses', icon: BookOpen },
      { label: 'Mi Progreso', path: '/progress', icon: TrendingUp },
      { label: 'Generar CV', path: '/generar-cv', icon: Award },
    ],
    admin: [
      { label: 'Dashboard', path: '/dashboard', icon: BarChart3 },
      { label: 'Gestión de Usuarios', path: '/users', icon: Users },
      { label: 'Seguimiento', path: '/seguimiento-participantes', icon: TrendingUp },
      { label: 'Cursos', path: '/courses', icon: BookOpen },
      { label: 'Campañas', path: '/gestion-campanas', icon: Megaphone },
    ],
    voluntario: [
      { label: 'Dashboard', path: '/dashboard', icon: BarChart3 },
      { label: 'Seguimiento', path: '/seguimiento-participantes', icon: TrendingUp },
      { label: 'Campañas', path: '/gestion-campanas', icon: Megaphone },
    ],
    docente: [
      { label: 'Dashboard', path: '/dashboard', icon: BarChart3 },
      { label: 'Portal Docente', path: '/teacher-portal', icon: GraduationCap },
      { label: 'Mis Cursos', path: '/courses', icon: BookOpen },
    ],
    empresa: [
      { label: 'Dashboard', path: '/dashboard', icon: BarChart3 },
      { label: 'Ofertas Laborales', path: '/courses', icon: BookOpen },
      { label: 'Candidatos', path: '/users', icon: Users },
    ],
  };

  const currentMenu = menuItems[userRole] || menuItems.participante;

  return (
    <div className="min-h-screen bg-background flex">
      {/* Sidebar */}
      <aside className={`${sidebarOpen ? 'w-64' : 'w-20'} bg-sidebar text-sidebar-foreground transition-all duration-300 flex flex-col`}>
        <div className="p-4 border-b border-sidebar-border">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center flex-shrink-0">
              <Leaf className="w-6 h-6 text-primary-foreground" />
            </div>
            {sidebarOpen && (
              <div>
                <h2 className="font-medium">Buen Vivir</h2>
                <p className="text-xs text-sidebar-foreground/70 capitalize">{userRole}</p>
              </div>
            )}
          </div>
        </div>

        <nav className="flex-1 p-4 space-y-2">
          {currentMenu.map((item) => (
            <Link
              key={item.path}
              to={item.path}
              className="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-sidebar-accent transition-colors"
            >
              <item.icon className="w-5 h-5 flex-shrink-0" />
              {sidebarOpen && <span>{item.label}</span>}
            </Link>
          ))}
        </nav>

        <div className="p-4 border-t border-sidebar-border space-y-2">
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-sidebar-accent transition-colors"
          >
            <Menu className="w-5 h-5 flex-shrink-0" />
            {sidebarOpen && <span>Contraer</span>}
          </button>
          <button
            onClick={onLogout}
            className="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-destructive transition-colors text-destructive-foreground"
          >
            <LogOut className="w-5 h-5 flex-shrink-0" />
            {sidebarOpen && <span>Salir</span>}
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-auto">
        {/* Header */}
        <header className="bg-card border-b border-border p-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl mb-1">
                {userRole === 'admin' && 'Panel de Administración'}
                {userRole === 'participante' && 'Mi Dashboard'}
                {userRole === 'voluntario' && 'Panel de Voluntario'}
                {userRole === 'docente' && 'Panel Docente'}
                {userRole === 'empresa' && 'Panel Empresarial'}
              </h1>
              <p className="text-muted-foreground">
                Bienvenido de nuevo, aquí está tu resumen
              </p>
            </div>
            <div className="flex items-center gap-4">
              <button className="relative p-2 rounded-lg hover:bg-secondary transition-colors">
                <Bell className="w-5 h-5" />
                <span className="absolute top-1 right-1 w-2 h-2 bg-destructive rounded-full"></span>
              </button>
              <button className="p-2 rounded-lg hover:bg-secondary transition-colors">
                <Settings className="w-5 h-5" />
              </button>
            </div>
          </div>
        </header>

        {/* Stats Grid */}
        <div className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            {currentStats.map((stat) => (
              <div key={stat.label} className="bg-card rounded-lg border border-border p-6">
                <div className="flex items-center justify-between mb-4">
                  <div className={`w-12 h-12 rounded-lg ${stat.color} flex items-center justify-center`}>
                    <stat.icon className="w-6 h-6 text-white" />
                  </div>
                  <span className="text-3xl">{stat.value}</span>
                </div>
                <h3 className="text-sm text-muted-foreground">{stat.label}</h3>
              </div>
            ))}
          </div>

          {/* Recent Activity / Quick Actions */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="bg-card rounded-lg border border-border p-6">
              <h2 className="text-xl mb-4">Actividad Reciente</h2>
              <div className="space-y-4">
                {[1, 2, 3, 4].map((i) => (
                  <div key={i} className="flex items-start gap-4 pb-4 border-b border-border last:border-0">
                    <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                      <BookOpen className="w-5 h-5 text-primary" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="mb-1">
                        {userRole === 'participante' && `Completaste la actividad ${i} del curso de Panadería`}
                        {userRole === 'admin' && `Nuevo participante registrado: Usuario ${i}`}
                        {userRole === 'voluntario' && `Sesión completada con participante ${i}`}
                        {userRole === 'docente' && `Nuevo material subido al curso ${i}`}
                        {userRole === 'empresa' && `Nueva aplicación a oferta laboral ${i}`}
                      </p>
                      <p className="text-sm text-muted-foreground">Hace {i} hora{i > 1 ? 's' : ''}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="bg-card rounded-lg border border-border p-6">
              <h2 className="text-xl mb-4">Acciones Rápidas</h2>
              <div className="space-y-3">
                <Link
                  to="/courses"
                  className="block p-4 rounded-lg border border-border hover:border-primary hover:bg-primary/5 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <BookOpen className="w-5 h-5 text-primary" />
                    <div>
                      <h3 className="mb-1">
                        {userRole === 'admin' && 'Crear Nuevo Curso'}
                        {userRole === 'participante' && 'Explorar Cursos'}
                        {userRole === 'voluntario' && 'Ver Mis Participantes'}
                        {userRole === 'docente' && 'Subir Material'}
                        {userRole === 'empresa' && 'Publicar Oferta'}
                      </h3>
                      <p className="text-sm text-muted-foreground">
                        {userRole === 'admin' && 'Agrega un nuevo curso al catálogo'}
                        {userRole === 'participante' && 'Descubre nuevos talleres y cursos'}
                        {userRole === 'voluntario' && 'Gestiona tus participantes asignados'}
                        {userRole === 'docente' && 'Comparte recursos con tus estudiantes'}
                        {userRole === 'empresa' && 'Encuentra el talento que necesitas'}
                      </p>
                    </div>
                  </div>
                </Link>

                <Link
                  to={userRole === 'admin' ? '/users' : '/progress'}
                  className="block p-4 rounded-lg border border-border hover:border-primary hover:bg-primary/5 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    {userRole === 'admin' ? <Users className="w-5 h-5 text-primary" /> : <BarChart3 className="w-5 h-5 text-primary" />}
                    <div>
                      <h3 className="mb-1">
                        {userRole === 'admin' && 'Gestionar Usuarios'}
                        {userRole === 'participante' && 'Ver Mi Progreso'}
                        {userRole === 'voluntario' && 'Registrar Asistencia'}
                        {userRole === 'docente' && 'Ver Estadísticas'}
                        {userRole === 'empresa' && 'Revisar Candidatos'}
                      </h3>
                      <p className="text-sm text-muted-foreground">
                        {userRole === 'admin' && 'Administra perfiles de participantes'}
                        {userRole === 'participante' && 'Revisa tu avance en los cursos'}
                        {userRole === 'voluntario' && 'Actualiza el progreso de participantes'}
                        {userRole === 'docente' && 'Analiza el rendimiento de tus cursos'}
                        {userRole === 'empresa' && 'Evalúa postulantes a tus ofertas'}
                      </p>
                    </div>
                  </div>
                </Link>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
