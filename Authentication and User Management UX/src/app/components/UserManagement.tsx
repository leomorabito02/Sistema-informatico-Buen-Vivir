import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Leaf, Search, Plus, Edit, Trash2, Eye, FileText, LogOut, X, User, Mail, Phone, Upload, CheckCircle, AlertCircle } from 'lucide-react';

interface UserManagementProps {
  onLogout: () => void;
}

export function UserManagement({ onLogout }: UserManagementProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [filterRole, setFilterRole] = useState('all');
  const [showModal, setShowModal] = useState(false);
  const [modalMode, setModalMode] = useState<'create' | 'edit' | 'view'>('create');
  const [selectedUser, setSelectedUser] = useState<any>(null);
  const [showStatusModal, setShowStatusModal] = useState(false);
  const [userToChangeStatus, setUserToChangeStatus] = useState<any>(null);
  const [usersList, setUsersList] = useState([
    { id: 1, name: 'María González', email: 'maria@email.com', phone: '+54 9 11 1234-5678', role: 'participante', status: 'Activo', courses: 3 },
    { id: 2, name: 'Juan Pérez', email: 'juan@email.com', phone: '+54 9 11 2345-6789', role: 'participante', status: 'Activo', courses: 5 },
    { id: 3, name: 'Ana Martínez', email: 'ana@email.com', phone: '+54 9 11 3456-7890', role: 'voluntario', status: 'Pendiente', courses: 0 },
    { id: 4, name: 'Carlos López', email: 'carlos@email.com', phone: '+54 9 11 4567-8901', role: 'docente', status: 'Activo', courses: 2 },
    { id: 5, name: 'Laura Rodríguez', email: 'laura@email.com', phone: '+54 9 11 5678-9012', role: 'participante', status: 'Activo', courses: 1 },
  ]);

  const filteredUsers = usersList.filter(user => {
    const matchesSearch = user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         user.email.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesRole = filterRole === 'all' || user.role === filterRole;
    return matchesSearch && matchesRole;
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

  const handleChangeStatus = (user: any) => {
    setUserToChangeStatus(user);
    setShowStatusModal(true);
  };

  const updateUserStatus = (newStatus: string) => {
    setUsersList(usersList.map(u => 
      u.id === userToChangeStatus.id ? { ...u, status: newStatus } : u
    ));
    setShowStatusModal(false);
    setUserToChangeStatus(null);
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="bg-primary/90 text-white border-b-2 border-primary p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link to="/dashboard" className="w-10 h-10 rounded-lg bg-white/20 flex items-center justify-center">
              <img src="/logo.png" alt="Buen Vivir" className="w-8 h-8 object-contain" />
            </Link>
            <div>
              <h1 className="text-2xl text-white">Gestión de Usuarios</h1>
              <p className="text-sm text-white/70">Administra perfiles de participantes y usuarios</p>
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
        {/* Filters and Actions */}
        <div className="bg-card rounded-lg border border-border p-6 mb-6">
          <div className="flex flex-col lg:flex-row gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="Buscar por nombre o email..."
              />
            </div>
            <select
              value={filterRole}
              onChange={(e) => setFilterRole(e.target.value)}
              className="px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
            >
              <option value="all">Todos los roles</option>
              <option value="participante">Participantes</option>
              <option value="voluntario">Voluntarios</option>
              <option value="docente">Docentes</option>
              <option value="empresa">Empresas</option>
            </select>
            <button
              onClick={handleCreateUser}
              className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity flex items-center gap-2 whitespace-nowrap"
            >
              <Plus className="w-5 h-5" />
              Nuevo Usuario
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
                  <th className="text-left p-4">Rol</th>
                  <th className="text-left p-4">Estado</th>
                  <th className="text-left p-4">Cursos</th>
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
                      <span className="inline-flex px-3 py-1 rounded-full text-sm bg-primary/10 text-primary capitalize">
                        {user.role}
                      </span>
                    </td>
                    <td className="p-4">
                      <div className="flex items-center gap-2">
                        <span className={`inline-flex px-3 py-1 rounded-full text-sm ${
                          user.status === 'Activo' ? 'bg-[#66bb6a]/10 text-[#2d7a3e]' : 'bg-[#f59e0b]/10 text-[#d97706]'
                        }`}>
                          {user.status}
                        </span>
                        <button
                          onClick={() => handleChangeStatus(user)}
                          className="p-1 rounded hover:bg-secondary transition-colors"
                          title="Cambiar estado"
                        >
                          {user.status === 'Activo' ? <AlertCircle className="w-4 h-4 text-amber-600" /> : <CheckCircle className="w-4 h-4 text-green-600" />}
                        </button>
                      </div>
                    </td>
                    <td className="p-4 text-muted-foreground">{user.courses}</td>
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
                          className="p-2 rounded-lg hover:bg-destructive/10 text-destructive transition-colors"
                          title="Eliminar"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                        <button
                          className="p-2 rounded-lg hover:bg-secondary transition-colors"
                          title="Ver log de auditoría"
                        >
                          <FileText className="w-4 h-4" />
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
                {modalMode === 'create' && 'Crear Nuevo Usuario'}
                {modalMode === 'edit' && 'Editar Usuario'}
                {modalMode === 'view' && 'Detalles del Usuario'}
              </h2>
              <button
                onClick={() => setShowModal(false)}
                className="p-2 rounded-lg hover:bg-secondary transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6">
              <form className="space-y-4">
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
                        placeholder="+54 9 11 1234-5678"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block mb-2">Rol</label>
                    <select
                      defaultValue={selectedUser?.role}
                      disabled={modalMode === 'view'}
                      className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                    >
                      <option value="participante">Participante</option>
                      <option value="voluntario">Voluntario</option>
                      <option value="docente">Docente</option>
                      <option value="empresa">Empresa</option>
                    </select>
                  </div>
                </div>

                {modalMode !== 'view' && (
                  <div>
                    <label className="block mb-2">Foto de Perfil</label>
                    <div className="border-2 border-dashed border-border rounded-lg p-8 text-center hover:border-primary transition-colors cursor-pointer">
                      <Upload className="w-8 h-8 text-muted-foreground mx-auto mb-2" />
                      <p className="text-sm text-muted-foreground mb-1">Haz clic o arrastra una imagen aquí</p>
                      <p className="text-xs text-muted-foreground">PNG, JPG hasta 5MB</p>
                    </div>
                  </div>
                )}

                <div>
                  <label className="block mb-2">Notas / Documentación</label>
                  <textarea
                    disabled={modalMode === 'view'}
                    className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50 min-h-[100px]"
                    placeholder="Información adicional sobre el participante..."
                  />
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
                      {modalMode === 'create' ? 'Crear Usuario' : 'Guardar Cambios'}
                    </button>
                  </div>
                )}
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Status Change Modal */}
      {showStatusModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-sm">
            <div className="border-b border-border p-6 flex items-center justify-between">
              <h2 className="text-xl font-semibold">Cambiar Estado</h2>
              <button
                onClick={() => setShowStatusModal(false)}
                className="p-2 rounded-lg hover:bg-secondary transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6 space-y-4">
              <p className="text-foreground">
                Usuario: <strong>{userToChangeStatus?.name}</strong>
              </p>
              <p className="text-sm text-muted-foreground">
                Estado actual: <strong>{userToChangeStatus?.status}</strong>
              </p>

              <div className="space-y-2 pt-4">
                <button
                  onClick={() => updateUserStatus('Activo')}
                  className="w-full px-4 py-2 rounded-lg border border-green-600 text-green-600 hover:bg-green-50 transition-colors flex items-center justify-center gap-2"
                >
                  <CheckCircle className="w-4 h-4" />
                  Activar Usuario
                </button>
                <button
                  onClick={() => updateUserStatus('Pendiente')}
                  className="w-full px-4 py-2 rounded-lg border border-amber-600 text-amber-600 hover:bg-amber-50 transition-colors flex items-center justify-center gap-2"
                >
                  <AlertCircle className="w-4 h-4" />
                  Poner en Pendiente
                </button>
              </div>

              <div className="flex justify-end gap-3 pt-4">
                <button
                  onClick={() => setShowStatusModal(false)}
                  className="px-6 py-2 rounded-lg border border-border hover:bg-secondary transition-colors"
                >
                  Cancelar
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
