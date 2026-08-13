import { Link } from 'react-router-dom';
import { BookOpen, Users, BarChart3, LogOut, Bell, Settings, GraduationCap, Award, TrendingUp, Megaphone, Clock, Target, CheckCircle, AlertCircle, Briefcase, Sparkles, Download, Filter } from 'lucide-react';
import { Button } from 'primereact/button';
import { useState } from 'react';
import { Card, CardHeader, CardBody, Toast, ToastState } from './base';

interface DashboardProps {
  userRole: 'admin' | 'participante' | 'voluntario' | 'empresa' | 'docente';
  onLogout: () => void;
}

// Mock data for recent activities
const recentActivities = [
];


const upcomingCourses = [
  { id: 1, name: 'Gastronomía Avanzada', startDate: '2026-06-01', duration: '8 semanas', progress: 85 },
  { id: 2, name: 'Gestión de Negocios', startDate: '2026-06-15', duration: '6 semanas', progress: 45 },
  { id: 3, name: 'Sostenibilidad Ambiental', startDate: '2026-07-01', duration: '4 semanas', progress: 0 },
];

// Evolución mensual mock de los indicadores principales del admin (HU-53)
const evolutionData: Record<string, number[]> = {
  'Participantes Activos': [10, 15, 17, 25, 28, 30],
  'Cursos Abiertos': [7, 8, 9, 10, 11, 12],
  'Cursos Completados': [3, 5, 6, 8, 9, 11],
  'Habilidades Desarrolladas': [5, 10, 15, 17, 20, 20],
  'Inserciones Laborales': [2, 4, 5, 7, 8, 9],
};
const evolutionMonths = ['Mar', 'Abr', 'May', 'Jun', 'Jul', 'Ago'];
const modules = ['Todos', 'Educación', 'Inserción Laboral', 'Promoción Humana', 'Campañas'];
const periods = ['Últimos 30 días', 'Últimos 3 meses', 'Últimos 6 meses', 'Este año'];

export function Dashboard({ userRole, onLogout }: DashboardProps) {
  const [chartIndicator, setChartIndicator] = useState('Participantes Activos');
  const [moduleFilter, setModuleFilter] = useState('Todos');
  const [periodFilter, setPeriodFilter] = useState('Últimos 6 meses');
  const [toast, setToast] = useState<ToastState | null>(null);

  const stats = {
    participante: [
      { label: 'Cursos Inscriptos ', value: '5', icon: BookOpen, color: 'bg-primary' },
      { label: 'Progreso Promedio', value: '68%', icon: TrendingUp, color: 'bg-secondary' },
      { label: 'Certificados', value: '2', icon: Award, color: 'bg-[#1D1D1B]' },
      { label: 'Horas de Estudio', value: '127', icon: GraduationCap, color: 'bg-primary/70' },
    ],
    admin: [
      { label: 'Participantes Activos', value: '30', icon: Users, color: 'bg-primary' },
      { label: 'Cursos Abiertos', value: '5', icon: BookOpen, color: 'bg-secondary' },
      { label: 'Cursos Completados', value: '10', icon: CheckCircle, color: 'bg-[#1D1D1B]' },
      { label: 'Habilidades Desarrolladas', value: '30', icon: Sparkles, color: 'bg-primary/70' },
      { label: 'Inserciones Laborales', value: '10', icon: Briefcase, color: 'bg-secondary/70' },
    ],
    voluntario: [
      { label: 'Participantes Asignados', value: '15', icon: Users, color: 'bg-primary' },
      { label: 'Sesiones Esta Semana', value: '8', icon: BookOpen, color: 'bg-secondary' },
      { label: 'Tarea Pendientes', value: '3', icon: BarChart3, color: 'bg-[#1D1D1B]' },
    ],
    docente: [
      { label: 'Mis Cursos', value: '3', icon: BookOpen, color: 'bg-primary' },
    ],
    empresa: [
      { label: 'Ofertas Activas', value: '7', icon: BookOpen, color: 'bg-primary' },
      { label: 'Candidatos', value: '23', icon: Users, color: 'bg-secondary' },
      { label: 'Contrataciones', value: '5', icon: Award, color: 'bg-[#1D1D1B]' },
    ],
  };

  const currentStats = stats[userRole] || stats.participante;

  const menuItems = {
    participante: [
      { label: 'Catálogo de Cursos', path: '/courses', icon: BookOpen },
      { label: 'Mi Progreso', path: '/progress', icon: TrendingUp },
      { label: 'Ofertas Laborales', path: '/ofertas-laborales', icon: Briefcase },
      { label: 'Generar CV', path: '/generar-cv', icon: Award },
    ],
    admin: [
      { label: 'Dashboard', path: '/dashboard', icon: BarChart3 },
      { label: 'Gestión de Usuarios', path: '/users', icon: Users },
      { label: 'Seguimiento', path: '/seguimiento-participantes', icon: TrendingUp },
      { label: 'Registro de Asistencia', path: '/asistencia-general', icon: CheckCircle },
      { label: 'Cursos', path: '/courses', icon: BookOpen },
      { label: 'Campañas', path: '/gestion-campanas', icon: Megaphone },
      { label: 'Inserción Laboral', path: '/insercion-laboral', icon: Award },
    ],
    voluntario: [
      { label: 'Cursos', path: '/courses', icon: BookOpen },
      { label: 'Seguimiento', path: '/seguimiento-participantes', icon: TrendingUp },
      { label: 'Registro de Asistencia', path: '/asistencia-general', icon: CheckCircle },
      { label: 'Ver Trabajos', path: '/insercion-laboral', icon: Award },
      { label: 'Gestión de Participantes', path: '/gestion-participantes', icon: Users },
    ],
    docente: [
      { label: 'Mis Cursos', path: '/teacher-portal', icon: GraduationCap },
    ],
    empresa: [
      { label: 'Ofertas Laborales', path: '/insercion-laboral', icon: BookOpen },
    ],
  };

  const currentMenu = menuItems[userRole] || menuItems.participante;

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Toast toast={toast} onClose={() => setToast(null)} />

      {/* Top Navigation Bar */}
      <header className="bg-primary/90 text-white border-b-2 border-primary p-4">
        <div className="flex items-center justify-between">
          {/* Logo and Brand */}
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-white/20 flex items-center justify-center flex-shrink-0">
              <img src="/logo.png" alt="Buen Vivir" className="w-8 h-8 object-contain" />
            </div>
            <div>
              <h2 className="font-medium text-white">Buen Vivir</h2>
              <p className="text-xs text-white/70 capitalize">{userRole}</p>
            </div>
          </div>

          {/* Navigation Links */}
          <nav className="flex items-center gap-2 flex-1 justify-center mx-8">
            {currentMenu.map((item) => (
              <Link
                key={item.path}
                to={item.path}
                className="px-3 py-2 rounded-lg hover:bg-white/10 transition-colors text-sm font-medium text-white"
              >
                {item.label}
              </Link>
            ))}
          </nav>

          {/* Right Side Actions */}
          <div className="flex items-center gap-4">
            <Link to="/notificaciones" className="relative p-2 rounded-lg hover:bg-white/10 transition-colors">
              <Bell className="w-5 h-5 text-white" />
              <span className="absolute top-1 right-1 w-2 h-2 bg-destructive rounded-full"></span>
            </Link>
            <Link to="/configuracion-notificaciones" className="p-2 rounded-lg hover:bg-white/10 transition-colors">
              <Settings className="w-5 h-5 text-white" />
            </Link>
            <Button
              onClick={onLogout}
              icon={<LogOut className="w-4 h-4 mr-2" />}
              label="Salir"
              className="px-4 py-2 rounded-lg bg-destructive text-destructive-foreground hover:opacity-90 transition-opacity border-none text-sm"
            />
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1 overflow-auto">


        {/* Stats Grid */}
        <div className="p-6">
          {userRole === 'admin' && (
            <div className="bg-card rounded-lg border border-border p-4 mb-6 flex flex-col sm:flex-row items-start sm:items-center gap-3 sm:gap-4">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Filter className="w-4 h-4" />
                Filtros:
              </div>
              <select
                value={periodFilter}
                onChange={(e) => setPeriodFilter(e.target.value)}
                className="px-3 py-2 rounded-lg border border-border bg-input-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              >
                {periods.map((p) => <option key={p} value={p}>{p}</option>)}
              </select>
              <select
                value={moduleFilter}
                onChange={(e) => setModuleFilter(e.target.value)}
                className="px-3 py-2 rounded-lg border border-border bg-input-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              >
                {modules.map((m) => <option key={m} value={m}>{m}</option>)}
              </select>
              <button
                onClick={() => setToast({ type: 'success', message: `Reporte del dashboard (${periodFilter}, módulo: ${moduleFilter}) exportado en PDF.` })}
                className="ml-auto px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors text-sm flex items-center gap-2"
              >
                <Download className="w-4 h-4" />
                Exportar PDF
              </button>
            </div>
          )}

          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-5 gap-6 mb-8">
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

          {/* Evolution Chart (HU-53) */}
          {userRole === 'admin' && (
            <Card className="mb-6">
              <CardHeader>
                <div className="flex items-center justify-between flex-wrap gap-3">
                  <h2 className="text-lg font-semibold flex items-center gap-2">
                    <BarChart3 size={20} className="text-primary" />
                    Evolución de Indicadores
                  </h2>
                  <select
                    value={chartIndicator}
                    onChange={(e) => setChartIndicator(e.target.value)}
                    className="px-3 py-2 rounded-lg border border-border bg-input-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                  >
                    {Object.keys(evolutionData).map((key) => <option key={key} value={key}>{key}</option>)}
                  </select>
                </div>
              </CardHeader>
              <CardBody>
                <div className="flex items-end justify-between gap-3 h-40">
                  {evolutionData[chartIndicator].map((value, idx) => {
                    const max = Math.max(...evolutionData[chartIndicator]);
                    const heightPct = Math.max(8, Math.round((value / max) * 100));
                    return (
                      <div key={idx} className="flex-1 flex flex-col items-center justify-end h-full gap-2">
                        <span className="text-xs text-muted-foreground">{value}</span>
                        <div
                          className="w-full bg-primary rounded-t-md transition-all"
                          style={{ height: `${heightPct}%` }}
                        />
                        <span className="text-xs text-muted-foreground">{evolutionMonths[idx]}</span>
                      </div>
                    );
                  })}
                </div>
              </CardBody>
            </Card>
          )}

          {/* Recent Activity & Upcoming Courses */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
            {/* Actividad Reciente */}
            <div className="lg:col-span-2">
              <Card>
                <CardHeader>
                  <h2 className="text-lg font-semibold flex items-center gap-2">
                    <Clock size={20} className="text-primary" />
                    Actividad Reciente
                  </h2>
                </CardHeader>
                <CardBody>
                  <div className="space-y-3">
                    {recentActivities
                      .filter((activity) => userRole !== 'admin' || moduleFilter === 'Todos' || activity.module === moduleFilter)
                      .map((activity) => (
                      <div key={activity.id} className="flex items-start gap-3 pb-3 border-b border-border last:border-0">
                        <div className={`w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0 ${
                          activity.type === 'success' ? 'bg-primary/10' :
                          activity.type === 'warning' ? 'bg-secondary/10' :
                          'bg-[#1D1D1B]/10'
                        }`}>
                          {activity.type === 'success' && <CheckCircle size={16} className="text-primary" />}
                          {activity.type === 'warning' && <AlertCircle size={16} className="text-secondary" />}
                          {activity.type === 'info' && <BookOpen size={16} className="text-[#1D1D1B]" />}
                        </div>
                        <div className="flex-1">
                          <p className="text-sm font-medium text-foreground">{activity.title}</p>
                          <p className="text-xs text-muted-foreground">{activity.timestamp}</p>
                        </div>
                      </div>
                    ))}
                  </div>
                </CardBody>
              </Card>
            </div>

            {/* Próximos Cursos (solo para participantes) */}
            {userRole === 'participante' && (
              <Card>
                <CardHeader>
                  <h2 className="text-lg font-semibold flex items-center gap-2">
                    <Target size={20} className="text-secondary" />
                    Próximos Cursos
                  </h2>
                </CardHeader>
                <CardBody>
                  <div className="space-y-3">
                    {upcomingCourses.map((course) => (
                      <div key={course.id} className="text-sm">
                        <p className="font-medium text-foreground mb-1">{course.name}</p>
                        <p className="text-xs text-muted-foreground mb-2">Inicia: {course.startDate}</p>
                        <div className="w-full bg-muted rounded-full h-1.5">
                          <div
                            className="bg-secondary h-1.5 rounded-full"
                            style={{ width: `${course.progress}%` }}
                          ></div>
                        </div>
                        <p className="text-xs text-muted-foreground mt-1">{course.progress}% completado</p>
                      </div>
                    ))}
                  </div>
                </CardBody>
              </Card>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}
