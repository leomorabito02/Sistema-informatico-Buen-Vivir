import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Leaf, Search, Plus, Edit, Trash2, Eye, LogOut, X, User, Mail, Phone } from 'lucide-react';

interface VolunteerUserManagementProps {
  onLogout: () => void;
}

export function VolunteerUserManagement({ onLogout }: VolunteerUserManagementProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [modalMode, setModalMode] = useState<'create' | 'edit' | 'view'>('create');
  const [selectedUser, setSelectedUser] = useState<any>(null);
  const [usersList, setUsersList] = useState([
    { id: 1, name: 'María González', email: 'maria@email.com', phone: '+54 9 11 1234-5678', role: 'participante', courses: 3 },
    { id: 2, name: 'Juan Pérez', email: 'juan@email.com', phone: '+54 9 11 2345-6789', role: 'participante', courses: 5 },
    { id: 3, name: 'Ana Martínez', email: 'ana@email.com', phone: '+54 9 11 3456-7890', role: 'participante', courses: 2 },
  ]);

  const filteredUsers = usersList.filter(user => {
    const matchesSearch = user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         user.email.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesSearch;
  });

  const handleCreateUser = () => {
    setModalMode('create');
    setSelectedUser(null);
    setShowModal(true);
  };

  const handleEditUser = (user: any) => {
    setModalMode('edit');
    setSelectedUser(user);
    setShowModal(true);
  };

  const handleViewUser = (user: any) => {
    setModalMode('view');
    setSelectedUser(user);
    setShowModal(true);
  };

  const handleDeleteUser = (userId: number) => {
    setUsersList(usersList.filter(u => u.id !== userId));
  };

  const handleSaveUser = () => {
    // Simular guardado de usuario
    setShowModal(false);
    setSelectedUser(null);
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="bg-primary/90 text-white border-b-2 border-primary p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link to="/dashboard" className="w-10 h-10 rounded-lg bg-white/20 flex items-center justify-center">
              <Leaf className="w-6 h-6 text-white" />
            </Link>
            <div>
              <h1 className="text-2xl text-white">Gestión de Participantes</h1>
              <p className="text-sm text-white/70">Administra los participantes bajo tu supervisión</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Link
              to="/dashboard"
              className="px-4 py-2 rounded-lg border border-white hover:bg-white/10 transition-colors text-white"
            >
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

      {/* Main Content */}
      <main className="p-6">
        {/* Search and Create */}
        <div className="bg-card rounded-lg border border-border p-6 mb-6">
          <div className="flex flex-col lg:flex-row gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="Buscar participantes..."
              />
            </div>
            <button
              onClick={handleCreateUser}
              className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity flex items-center gap-2 whitespace-nowrap"
            >
              <Plus className="w-5 h-5" />
              Nuevo Participante
            </button>
          </div>
        </div>

        {/* Users Table */}
        <div className="bg-card rounded-lg border border-border overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-muted">
                <tr>
                  <th className="text-left p-4">Nombre</th>
                  <th className="text-left p-4">Email</th>
                  <th className="text-left p-4">Teléfono</th>
                  <th className="text-left p-4">Cursos Inscritos</th>
                  <th className="text-left p-4">Acciones</th>
                </tr>
              </thead>
              <tbody>
                {filteredUsers.map((user) => (
                  <tr key={user.id} className="border-t border-border hover:bg-muted/50 transition-colors">
                    <td className="p-4">
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
                          <User className="w-5 h-5 text-primary" />
                        </div>
                        <span>{user.name}</span>
                      </div>
                    </td>
                    <td className="p-4 text-muted-foreground">{user.email}</td>
                    <td className="p-4 text-muted-foreground">{user.phone}</td>
                    <td className="p-4">
                      <span className="inline-flex px-3 py-1 rounded-full text-sm bg-primary/10 text-primary">
                        {user.courses}
                      </span>
                    </td>
                    <td className="p-4">
                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => handleViewUser(user)}
                          className="p-2 rounded-lg hover:bg-secondary transition-colors"
                          title="Ver detalles"
                        >
                          <Eye className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleEditUser(user)}
                          className="p-2 rounded-lg hover:bg-secondary transition-colors"
                          title="Editar"
                        >
                          <Edit className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleDeleteUser(user.id)}
                          className="p-2 rounded-lg hover:bg-destructive/10 text-destructive transition-colors"
                          title="Eliminar"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </main>

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-card border-b border-border p-6 flex items-center justify-between">
              <h2 className="text-2xl">
                {modalMode === 'create' && 'Registrar Nuevo Participante'}
                {modalMode === 'edit' && 'Editar Participante'}
                {modalMode === 'view' && 'Detalles del Participante'}
              </h2>
              <button
                onClick={() => setShowModal(false)}
                className="p-2 rounded-lg hover:bg-secondary transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6">
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  handleSaveUser();
                }}
                className="space-y-4"
              >
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block mb-2">Nombre Completo</label>
                    <div className="relative">
                      <User className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                      <input
                        type="text"
                        defaultValue={selectedUser?.name}
                        disabled={modalMode === 'view'}
                        className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                        placeholder="Juan Pérez"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block mb-2">Email</label>
                    <div className="relative">
                      <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                      <input
                        type="email"
                        defaultValue={selectedUser?.email}
                        disabled={modalMode === 'view'}
                        className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                        placeholder="email@ejemplo.com"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block mb-2">Teléfono</label>
                    <div className="relative">
                      <Phone className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                      <input
                        type="tel"
                        defaultValue={selectedUser?.phone}
                        disabled={modalMode === 'view'}
                        className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                        placeholder="+54 9 11 XXXX-XXXX"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block mb-2">Cursos Inscritos</label>
                    <input
                      type="number"
                      defaultValue={selectedUser?.courses || 0}
                      disabled={modalMode === 'view'}
                      className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                      placeholder="0"
                    />
                  </div>
                </div>

                <div className="bg-secondary rounded-lg p-4">
                  <p className="text-sm text-muted-foreground">
                    <strong className="text-foreground">Información:</strong> Los participantes registrados por aquí podrán acceder 
                    a los cursos y seguimiento de asistencia.
                  </p>
                </div>

                {modalMode !== 'view' && (
                  <div className="flex justify-end gap-3 pt-4">
                    <button
                      type="button"
                      onClick={() => setShowModal(false)}
                      className="px-6 py-3 rounded-lg border border-border hover:bg-secondary transition-colors"
                    >
                      Cancelar
                    </button>
                    <button
                      type="submit"
                      className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
                    >
                      {modalMode === 'create' ? 'Registrar Participante' : 'Guardar Cambios'}
                    </button>
                  </div>
                )}
              </form>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
