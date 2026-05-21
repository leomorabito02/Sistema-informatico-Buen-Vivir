import React, { useState } from 'react';
import { Bell, X, CheckCircle, AlertCircle, BookOpen, Briefcase, Heart, Settings } from 'lucide-react';
import { Card, CardHeader, CardBody } from './base';

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
    actionUrl: '/certificates/panaderia',
  },
  {
    id: 2,
    title: 'Nueva Oferta Laboral',
    message: 'Se publicó una oferta de pastelería que coincide con tu perfil.',
    type: 'info',
    timestamp: 'Hace 5 horas',
    read: false,
    category: 'job',
    actionUrl: '/ofertas-laborales/45',
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

const notificationStats = [
  { label: 'No leídas', value: 2, color: 'bg-blue-100 text-blue-700' },
  { label: 'Esta Semana', value: 5, color: 'bg-green-100 text-green-700' },
  { label: 'Alertas', value: 1, color: 'bg-red-100 text-red-700' },
];

interface NotificationCenterProps {
  onClose?: () => void;
}

export function NotificationCenter({ onClose }: NotificationCenterProps) {
  const [notifications, setNotifications] = useState(mockNotifications);
  const [selectedFilter, setSelectedFilter] = useState<'all' | 'unread' | 'course' | 'job' | 'campaign'>('all');

  const filteredNotifications = notifications.filter((notif) => {
    if (selectedFilter === 'unread') return !notif.read;
    if (selectedFilter === 'course') return notif.category === 'course';
    if (selectedFilter === 'job') return notif.category === 'job';
    if (selectedFilter === 'campaign') return notif.category === 'campaign';
    return true;
  });

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
        return <CheckCircle size={20} className="text-green-600" />;
      case 'warning':
        return <AlertCircle size={20} className="text-yellow-600" />;
      case 'alert':
        return <AlertCircle size={20} className="text-red-600" />;
      default:
        return <Bell size={20} className="text-blue-600" />;
    }
  };

  const getCategoryIcon = (category: Notification['category']) => {
    switch (category) {
      case 'course':
        return <BookOpen size={16} className="text-blue-600" />;
      case 'job':
        return <Briefcase size={16} className="text-green-600" />;
      case 'campaign':
        return <Heart size={16} className="text-pink-600" />;
      default:
        return <Settings size={16} className="text-gray-600" />;
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 pb-6">
      {/* Header */}
      <div className="bg-white border-b border-gray-200 sticky top-0 z-40">
        <div className="max-w-6xl mx-auto px-6 py-6">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 rounded-lg bg-blue-100 flex items-center justify-center">
                <Bell size={24} className="text-blue-600" />
              </div>
              <div>
                <h1 className="text-3xl font-bold text-gray-900">Centro de Notificaciones</h1>
                <p className="text-gray-600">Gestiona todas tus notificaciones en un solo lugar</p>
              </div>
            </div>
            {onClose && (
              <button
                onClick={onClose}
                className="p-2 rounded-lg hover:bg-gray-100 transition-colors"
              >
                <X size={24} className="text-gray-600" />
              </button>
            )}
          </div>

          {/* Stats */}
          <div className="grid grid-cols-3 gap-4 mb-6">
            {notificationStats.map((stat) => (
              <div key={stat.label} className={`px-4 py-2 rounded-lg ${stat.color}`}>
                <p className="font-bold">{stat.value}</p>
                <p className="text-xs">{stat.label}</p>
              </div>
            ))}
          </div>

          {/* Filters */}
          <div className="flex gap-2 overflow-x-auto pb-2">
            {['all', 'unread', 'course', 'job', 'campaign'].map((filter) => (
              <button
                key={filter}
                onClick={() => setSelectedFilter(filter as any)}
                className={`px-4 py-2 rounded-lg whitespace-nowrap transition-colors ${
                  selectedFilter === filter
                    ? 'bg-blue-600 text-white'
                    : 'bg-white text-gray-700 border border-gray-200 hover:border-blue-600'
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
        </div>

        {/* Mark All As Read */}
        {notifications.some((n) => !n.read) && (
          <div className="bg-blue-50 border-t border-blue-200 px-6 py-3">
            <button
              onClick={markAllAsRead}
              className="text-blue-700 text-sm font-medium hover:text-blue-800"
            >
              Marcar todas como leídas
            </button>
          </div>
        )}
      </div>

      {/* Notifications List */}
      <div className="max-w-6xl mx-auto px-6 pt-6">
        <div className="space-y-4">
          {filteredNotifications.length > 0 ? (
            filteredNotifications.map((notification) => (
              <Card
                key={notification.id}
                className={`${
                  !notification.read ? 'border-l-4 border-l-blue-600 bg-blue-50' : ''
                }`}
              >
                <CardBody className="p-4">
                  <div className="flex items-start gap-4">
                    <div className="flex-shrink-0 mt-1">{getNotificationIcon(notification.type)}</div>

                    <div className="flex-1 min-w-0">
                      <div className="flex items-start justify-between gap-4">
                        <div className="flex-1">
                          <h3 className={`font-semibold ${!notification.read ? 'text-gray-900' : 'text-gray-700'}`}>
                            {notification.title}
                          </h3>
                          <p className="text-gray-600 text-sm mt-1">{notification.message}</p>
                          <div className="flex items-center gap-4 mt-3">
                            <span className="text-xs text-gray-500">{notification.timestamp}</span>
                            <div className="flex items-center gap-2">
                              {getCategoryIcon(notification.category)}
                              <span className="text-xs text-gray-600">
                                {notification.category === 'course' && 'Curso'}
                                {notification.category === 'job' && 'Oferta Laboral'}
                                {notification.category === 'campaign' && 'Campaña'}
                                {notification.category === 'system' && 'Sistema'}
                              </span>
                            </div>
                          </div>
                        </div>
                      </div>

                      {notification.actionUrl && (
                        <div className="mt-3">
                          <a
                            href={notification.actionUrl}
                            className="text-sm text-blue-600 hover:text-blue-800 font-medium"
                          >
                            Ver más →
                          </a>
                        </div>
                      )}
                    </div>

                    <div className="flex-shrink-0 flex items-center gap-2">
                      {!notification.read && (
                        <button
                          onClick={() => markAsRead(notification.id)}
                          className="px-3 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
                        >
                          Marcar leída
                        </button>
                      )}
                      <button
                        onClick={() => deleteNotification(notification.id)}
                        className="p-1 text-gray-400 hover:text-gray-600 transition-colors"
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
              <Bell size={48} className="mx-auto text-gray-300 mb-4" />
              <p className="text-gray-500 text-lg">No hay notificaciones</p>
              <p className="text-gray-400 text-sm">Volveremos aquí cuando tengas nuevas notificaciones</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
