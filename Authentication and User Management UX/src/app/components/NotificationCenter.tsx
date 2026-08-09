import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Bell, X, CheckCircle, AlertCircle, BookOpen, Briefcase, Heart, Settings, LogOut, ArrowLeft } from 'lucide-react';
import { Card, CardBody } from './base';

interface Notification {
  id: number;
  title: string;
  message: string;
  type: 'success' | 'warning' | 'info' | 'alert';
  timestamp: string;
  read: boolean;
  category: 'course' | 'job' | 'campaign' | 'system';
  actionUrl?: string;
}

const mockNotifications: Notification[] = [
  {
    id: 1,
    title: '¡Congratulaciones!',
    message: 'Completaste el curso de Panadería Artesanal. Descarga tu certificado.',
    type: 'success',
    timestamp: 'Hace 2 horas',
    read: false,
    category: 'course',
    actionUrl: '/progress',
  },
  {
    id: 2,
    title: 'Nueva Oferta Laboral',
    message: 'Se publicó una oferta de pastelería que coincide con tu perfil.',
    type: 'info',
    timestamp: 'Hace 5 horas',
    read: false,
    category: 'job',
    actionUrl: '/ofertas-laborales',
  },
  {
    id: 3,
    title: 'Recordatorio de Clase',
    message: 'Tu clase de Costura y Diseño comienza en 1 hora.',
    type: 'alert',
    timestamp: 'Hace 1 hora',
    read: true,
    category: 'course',
  },
  {
    id: 4,
    title: 'Campaña Especial',
    message: 'Nueva campaña de empleabilidad: "Encuentra tu Primer Trabajo"',
    type: 'info',
    timestamp: 'Ayer',
    read: true,
    category: 'campaign',
  },
  {
    id: 5,
    title: '¡Alerta!',
    message: 'Te falta enviar la tarea de la Actividad 3 en Agricultura.',
    type: 'warning',
    timestamp: 'Ayer',
    read: true,
    category: 'course',
  },
  {
    id: 6,
    title: 'Nuevo Voluntario Asignado',
    message: 'María Rodríguez será tu voluntaria de acompañamiento.',
    type: 'success',
    timestamp: 'Hace 2 días',
    read: true,
    category: 'system',
  },
];

interface NotificationCenterProps {
  onLogout: () => void;
}

export function NotificationCenter({ onLogout }: NotificationCenterProps) {
  const [notifications, setNotifications] = useState(mockNotifications);
  const [selectedFilter, setSelectedFilter] = useState<'all' | 'unread' | 'course' | 'job' | 'campaign'>('all');

  const filteredNotifications = notifications.filter((notif) => {
    if (selectedFilter === 'unread') return !notif.read;
    if (selectedFilter === 'course') return notif.category === 'course';
    if (selectedFilter === 'job') return notif.category === 'job';
    if (selectedFilter === 'campaign') return notif.category === 'campaign';
    return true;
  });

  const notificationStats = [
    { label: 'No leídas', value: notifications.filter(n => !n.read).length },
    { label: 'Esta Semana', value: notifications.length },
    { label: 'Alertas', value: notifications.filter(n => n.type === 'alert').length },
  ];

  const markAsRead = (id: number) => {
    setNotifications(notifications.map((n) => (n.id === id ? { ...n, read: true } : n)));
  };

  const deleteNotification = (id: number) => {
    setNotifications(notifications.filter((n) => n.id !== id));
  };

  const markAllAsRead = () => {
    setNotifications(notifications.map((n) => ({ ...n, read: true })));
  };

  const getNotificationIcon = (type: Notification['type']) => {
    switch (type) {
      case 'success':
        return <CheckCircle size={20} className="text-primary" />;
      case 'warning':
        return <AlertCircle size={20} className="text-secondary" />;
      case 'alert':
        return <AlertCircle size={20} className="text-destructive" />;
      default:
        return <Bell size={20} className="text-[#1D1D1B]" />;
    }
  };

  const getCategoryIcon = (category: Notification['category']) => {
    switch (category) {
      case 'course':
        return <BookOpen size={16} className="text-primary" />;
      case 'job':
        return <Briefcase size={16} className="text-primary" />;
      case 'campaign':
        return <Heart size={16} className="text-secondary" />;
      default:
        return <Settings size={16} className="text-muted-foreground" />;
    }
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
              <h1 className="text-2xl flex items-center gap-2">
                <Bell className="w-6 h-6 text-primary" />
                Centro de Notificaciones
              </h1>
              <p className="text-sm text-muted-foreground">Gestioná todas tus notificaciones en un solo lugar</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Link
              to="/configuracion-notificaciones"
              className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors flex items-center gap-2"
            >
              <Settings className="w-4 h-4" />
              Configurar
            </Link>
            <Link to="/dashboard" className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors flex items-center gap-2">
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

      <main className="p-6 max-w-5xl mx-auto">
        {/* Stats */}
        <div className="grid grid-cols-3 gap-4 mb-6">
          {notificationStats.map((stat) => (
            <div key={stat.label} className="bg-card rounded-lg border border-border px-4 py-3">
              <p className="text-2xl font-semibold text-primary">{stat.value}</p>
              <p className="text-xs text-muted-foreground">{stat.label}</p>
            </div>
          ))}
        </div>

        {/* Filters */}
        <div className="flex gap-2 overflow-x-auto pb-2 mb-4">
          {['all', 'unread', 'course', 'job', 'campaign'].map((filter) => (
            <button
              key={filter}
              onClick={() => setSelectedFilter(filter as any)}
              className={`px-4 py-2 rounded-lg whitespace-nowrap transition-colors text-sm ${
                selectedFilter === filter
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-card text-foreground border border-border hover:border-primary'
              }`}
            >
              {filter === 'all' && 'Todas'}
              {filter === 'unread' && 'No Leídas'}
              {filter === 'course' && 'Cursos'}
              {filter === 'job' && 'Empleabilidad'}
              {filter === 'campaign' && 'Campañas'}
            </button>
          ))}
        </div>

        {/* Mark All As Read */}
        {notifications.some((n) => !n.read) && (
          <div className="bg-primary/5 border border-primary/20 rounded-lg px-4 py-3 mb-6">
            <button
              onClick={markAllAsRead}
              className="text-primary text-sm font-medium hover:underline"
            >
              Marcar todas como leídas
            </button>
          </div>
        )}

        {/* Notifications List */}
        <div className="space-y-4">
          {filteredNotifications.length > 0 ? (
            filteredNotifications.map((notification) => (
              <Card
                key={notification.id}
                className={`bg-card ${!notification.read ? 'border-l-4 border-l-primary' : ''}`}
              >
                <CardBody className="p-4">
                  <div className="flex items-start gap-4">
                    <div className="flex-shrink-0 mt-1">{getNotificationIcon(notification.type)}</div>

                    <div className="flex-1 min-w-0">
                      <h3 className={`font-semibold ${!notification.read ? 'text-foreground' : 'text-muted-foreground'}`}>
                        {notification.title}
                      </h3>
                      <p className="text-muted-foreground text-sm mt-1">{notification.message}</p>
                      <div className="flex items-center gap-4 mt-3">
                        <span className="text-xs text-muted-foreground">{notification.timestamp}</span>
                        <div className="flex items-center gap-2">
                          {getCategoryIcon(notification.category)}
                          <span className="text-xs text-muted-foreground">
                            {notification.category === 'course' && 'Curso'}
                            {notification.category === 'job' && 'Oferta Laboral'}
                            {notification.category === 'campaign' && 'Campaña'}
                            {notification.category === 'system' && 'Sistema'}
                          </span>
                        </div>
                      </div>

                      {notification.actionUrl && (
                        <div className="mt-3">
                          <Link
                            to={notification.actionUrl}
                            className="text-sm text-primary hover:underline font-medium"
                          >
                            Ver más →
                          </Link>
                        </div>
                      )}
                    </div>

                    <div className="flex-shrink-0 flex items-center gap-2">
                      {!notification.read && (
                        <button
                          onClick={() => markAsRead(notification.id)}
                          className="px-3 py-1 text-xs bg-primary text-primary-foreground rounded hover:opacity-90 transition-opacity"
                        >
                          Marcar leída
                        </button>
                      )}
                      <button
                        onClick={() => deleteNotification(notification.id)}
                        className="p-1 text-muted-foreground hover:text-foreground transition-colors"
                      >
                        <X size={16} />
                      </button>
                    </div>
                  </div>
                </CardBody>
              </Card>
            ))
          ) : (
            <div className="text-center py-12">
              <Bell size={48} className="mx-auto text-muted-foreground/30 mb-4" />
              <p className="text-muted-foreground text-lg">No hay notificaciones</p>
              <p className="text-muted-foreground/70 text-sm">Volveremos aquí cuando tengas nuevas notificaciones</p>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
