import { Link } from 'react-router-dom';
import { Leaf, BookOpen, Users, BarChart3, LogOut, Menu, Bell, Settings, GraduationCap, Award, TrendingUp, Megaphone, Clock, Target, CheckCircle, AlertCircle, Phone, FileText } from 'lucide-react';
import { Button } from 'primereact/button';
import { useState } from 'react';
import { Card, CardHeader, CardBody } from './base';

interface DashboardProps {
  userRole: 'admin' | 'participante' | 'voluntario' | 'empresa' | 'docente';
  onLogout: () => void;
}

// Mock data for recent activities
const recentActivities = [
  { id: 1, title: 'Cursó completado: Gastronomía Básica', timestamp: 'Hace 2 horas', type: 'success' },
  { id: 2, title: 'Nueva oferta laboral disponible en tu área', timestamp: 'Ayer', type: 'info' },
  { id: 3, title: 'Recordatorio: Tarea vencida en Agricultura', timestamp: 'Ayer', type: 'warning' },
  { id: 4, title: 'Certificado descargado: Oficios', timestamp: 'Hace 3 días', type: 'success' },
];

const upcomingCourses = [
  { id: 1, name: 'Gastronomía Avanzada', startDate: '2026-06-01', duration: '8 semanas', progress: 85 },
  { id: 2, name: 'Gestión de Negocios', startDate: '2026-06-15', duration: '6 semanas', progress: 45 },
  { id: 3, name: 'Sostenibilidad Ambiental', startDate: '2026-07-01', duration: '4 semanas', progress: 0 },
];

export function Dashboard({ userRole, onLogout }: DashboardProps) {
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
      { label: 'Inserción Laboral', path: '/insercion-laboral', icon: Award },
    ],
    voluntario: [
      { label: 'Cursos', path: '/courses', icon: BookOpen },
      { label: 'Registrar Asistencia', path: '/seguimiento-participantes', icon: TrendingUp },
      { label: 'Ver Trabajos', path: '/insercion-laboral', icon: Award },
      { label: 'Gestión de Participantes', path: '/gestion-participantes', icon: Users },
    ],
    docente: [
      { label: 'Portal Docente', path: '/teacher-portal', icon: GraduationCap },
      { label: 'Mis Cursos', path: '/courses', icon: BookOpen },
    ],
    empresa: [
      { label: 'Ofertas Laborales', path: '/insercion-laboral', icon: BookOpen },
    ],
  };

  const currentMenu = menuItems[userRole] || menuItems.participante;

  return (
    <div className="min-h-screen bg-background flex flex-col">
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
            <Button icon={<Bell className="w-5 h-5 text-white" />} className="relative p-2 rounded-lg hover:bg-white/10 transition-colors p-button-text">
              <span className="absolute top-1 right-1 w-2 h-2 bg-destructive rounded-full"></span>
            </Button>
            <Button icon={<Settings className="w-5 h-5 text-white" />} className="p-2 rounded-lg hover:bg-white/10 transition-colors p-button-text" />
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

          {/* Recent Activity & Upcoming Courses */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
            {/* Actividad Reciente */}
            <div className="lg:col-span-2">
              <Card>
                <CardHeader>
                  <h2 className="text-lg font-semibold flex items-center gap-2">
                    <Clock size={20} className="text-blue-600" />
                    Actividad Reciente
                  </h2>
                </CardHeader>
                <CardBody>
                  <div className="space-y-3">
                    {recentActivities.map((activity) => (
                      <div key={activity.id} className="flex items-start gap-3 pb-3 border-b border-gray-100 last:border-0">
                        <div className={`w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0 ${
                          activity.type === 'success' ? 'bg-green-100' :
                          activity.type === 'warning' ? 'bg-yellow-100' :
                          'bg-blue-100'
                        }`}>
                          {activity.type === 'success' && <CheckCircle size={16} className="text-green-600" />}
                          {activity.type === 'warning' && <AlertCircle size={16} className="text-yellow-600" />}
                          {activity.type === 'info' && <BookOpen size={16} className="text-blue-600" />}
                        </div>
                        <div className="flex-1">
                          <p className="text-sm font-medium text-gray-900">{activity.title}</p>
                          <p className="text-xs text-gray-500">{activity.timestamp}</p>
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
                    <Target size={20} className="text-purple-600" />
                    Próximos Cursos
                  </h2>
                </CardHeader>
                <CardBody>
                  <div className="space-y-3">
                    {upcomingCourses.map((course) => (
                      <div key={course.id} className="text-sm">
                        <p className="font-medium text-gray-900 mb-1">{course.name}</p>
                        <p className="text-xs text-gray-500 mb-2">Inicia: {course.startDate}</p>
                        <div className="w-full bg-gray-200 rounded-full h-1.5">
                          <div
                            className="bg-blue-600 h-1.5 rounded-full"
                            style={{ width: `${course.progress}%` }}
                          ></div>
                        </div>
                        <p className="text-xs text-gray-500 mt-1">{course.progress}% completado</p>
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
