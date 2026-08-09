import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Search, Plus, Edit, Trash2, Eye, FileText, LogOut, X, User, Mail, Phone, Upload, CheckCircle, AlertCircle, IdCard, MapPin, Calendar, Lock, ArrowLeft } from 'lucide-react';
import { Toast, ToastState } from './base';

interface UserManagementProps {
  onLogout: () => void;
}

const GMAIL_REGEX = /^[^\s@]+@gmail\.com$/i;
const PASSWORD_REGEX = /^(?=.*[A-Z])(?=.*[^A-Za-z0-9]).{8,}$/;

const emptyFormData = {
  name: '',
  email: '',
  phone: '',
  role: 'participante',
  password: '',
  dni: '',
  address: '',
  birthDate: '',
};

export function UserManagement({ onLogout }: UserManagementProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [filterRole, setFilterRole] = useState('all');
  const [showModal, setShowModal] = useState(false);
  const [modalMode, setModalMode] = useState<'create' | 'edit' | 'view'>('create');
  const [selectedUser, setSelectedUser] = useState<any>(null);
  const [showStatusModal, setShowStatusModal] = useState(false);
  const [userToChangeStatus, setUserToChangeStatus] = useState<any>(null);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [userToDelete, setUserToDelete] = useState<any>(null);
  const [formData, setFormData] = useState(emptyFormData);
  const [formError, setFormError] = useState('');
  const [toast, setToast] = useState<ToastState | null>(null);
  const [usersList, setUsersList] = useState([
    { id: 1, name: 'María González', email: 'maria@gmail.com', phone: '+54 9 11 1234-5678', role: 'participante', status: 'Activo', courses: 3, dni: '30111222', address: 'Belgrano 123, Alta Gracia', birthDate: '1997-03-12' },
    { id: 2, name: 'Juan Pérez', email: 'juan@gmail.com', phone: '+54 9 11 2345-6789', role: 'participante', status: 'Activo', courses: 5, dni: '28555444', address: 'San Martín 456, Alta Gracia', birthDate: '1990-07-22' },
    { id: 3, name: 'Ana Martínez', email: 'ana@gmail.com', phone: '+54 9 11 3456-7890', role: 'voluntario', status: 'Pendiente', courses: 0, dni: '33222111', address: 'Sarmiento 789, Alta Gracia', birthDate: '2000-01-05' },
    { id: 4, name: 'Carlos López', email: 'carlos@gmail.com', phone: '+54 9 11 4567-8901', role: 'docente', status: 'Activo', courses: 2, dni: '27888999', address: 'Rivadavia 321, Alta Gracia', birthDate: '1985-11-30' },
    { id: 5, name: 'Laura Rodríguez', email: 'laura@gmail.com', phone: '+54 9 11 5678-9012', role: 'participante', status: 'Activo', courses: 1, dni: '31444555', address: 'Mitre 654, Alta Gracia', birthDate: '1995-09-18' },
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
    setFormData(emptyFormData);
    setFormError('');
    setShowModal(true);
  };

  const handleEditUser = (user: any) => {
    setModalMode('edit');
    setSelectedUser(user);
    setFormData({
      name: user.name, email: user.email, phone: user.phone, role: user.role,
      password: '', dni: user.dni || '', address: user.address || '', birthDate: user.birthDate || '',
    });
    setFormError('');
    setShowModal(true);
  };

  const handleViewUser = (user: any) => {
    setModalMode('view');
    setSelectedUser(user);
    setFormData({
      name: user.name, email: user.email, phone: user.phone, role: user.role,
      password: '', dni: user.dni || '', address: user.address || '', birthDate: user.birthDate || '',
    });
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
    setToast({
      type: 'success',
      message: `${userToChangeStatus.name} fue notificado/a: cuenta ${newStatus === 'Activo' ? 'aprobada' : 'puesta en pendiente'}.`,
    });
    setUserToChangeStatus(null);
  };

  const handleAskDelete = (user: any) => {
    setUserToDelete(user);
    setShowDeleteModal(true);
  };

  const confirmDeleteUser = () => {
    setUsersList(usersList.filter(u => u.id !== userToDelete.id));
    setShowDeleteModal(false);
    setToast({ type: 'success', message: `${userToDelete.name} fue eliminado del sistema (baja lógica registrada en auditoría).` });
    setUserToDelete(null);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!formData.name.trim() || !formData.email.trim() || !formData.phone.trim() ||
        !formData.dni.trim() || !formData.address.trim() || !formData.birthDate.trim()) {
      setFormError('Completá todos los campos obligatorios.');
      return;
    }

    if (!GMAIL_REGEX.test(formData.email.trim())) {
      setFormError('El email debe tener dominio @gmail.com.');
      return;
    }

    const duplicated = usersList.some(
      u => u.email.toLowerCase() === formData.email.trim().toLowerCase() && u.id !== selectedUser?.id
    );
    if (duplicated) {
      setFormError('Ya existe un usuario registrado con ese email.');
      return;
    }

    if (modalMode === 'create') {
      if (!PASSWORD_REGEX.test(formData.password)) {
        setFormError('La contraseña debe tener al menos 8 caracteres, una mayúscula y un carácter especial.');
        return;
      }
      const newUser = {
        id: Date.now(),
        name: formData.name.trim(),
        email: formData.email.trim(),
        phone: formData.phone.trim(),
        role: formData.role,
        status: 'Activo',
        courses: 0,
        dni: formData.dni.trim(),
        address: formData.address.trim(),
        birthDate: formData.birthDate,
      };
      setUsersList([newUser, ...usersList]);
      setToast({ type: 'success', message: 'Usuario creado correctamente.' });
    } else if (modalMode === 'edit') {
      setUsersList(usersList.map(u => u.id === selectedUser.id ? {
        ...u,
        name: formData.name.trim(),
        email: formData.email.trim(),
        phone: formData.phone.trim(),
        role: formData.role,
        dni: formData.dni.trim(),
        address: formData.address.trim(),
        birthDate: formData.birthDate,
      } : u));
      setToast({ type: 'success', message: 'Cambios guardados correctamente' });
    }

    setFormError('');
    setShowModal(false);
  };

  return (
    <div className="min-h-screen bg-background">
      <Toast toast={toast} onClose={() => setToast(null)} />

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
              className="px-4 py-2 rounded-lg border border-white hover:bg-white/10 transition-colors text-white flex items-center gap-2"
            >
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
                          user.status === 'Activo' ? 'bg-primary/10 text-primary' : 'bg-secondary/10 text-secondary'
                        }`}>
                          {user.status}
                        </span>
                        <button
                          onClick={() => handleChangeStatus(user)}
                          className="p-1 rounded hover:bg-secondary transition-colors"
                          title="Aprobar / cambiar estado"
                        >
                          {user.status === 'Activo' ? <AlertCircle className="w-4 h-4 text-secondary" /> : <CheckCircle className="w-4 h-4 text-primary" />}
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
                          onClick={() => handleAskDelete(user)}
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

      {/* Create / Edit / View Modal */}
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
              <form onSubmit={handleSubmit} className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block mb-2">Nombre Completo</label>
                    <div className="relative">
                      <User className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                      <input
                        type="text"
                        value={formData.name}
                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                        disabled={modalMode === 'view'}
                        className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                        placeholder="Juan Pérez"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block mb-2">Email (dominio @gmail.com)</label>
                    <div className="relative">
                      <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                      <input
                        type="email"
                        value={formData.email}
                        onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                        disabled={modalMode === 'view'}
                        className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                        placeholder="usuario@gmail.com"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block mb-2">Teléfono</label>
                    <div className="relative">
                      <Phone className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                      <input
                        type="tel"
                        value={formData.phone}
                        onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                        disabled={modalMode === 'view'}
                        className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                        placeholder="+54 9 11 1234-5678"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block mb-2">Rol</label>
                    <select
                      value={formData.role}
                      onChange={(e) => setFormData({ ...formData, role: e.target.value })}
                      disabled={modalMode === 'view'}
                      className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                    >
                      <option value="participante">Participante</option>
                      <option value="voluntario">Voluntario</option>
                      <option value="docente">Docente</option>
                      <option value="empresa">Empresa</option>
                    </select>
                  </div>

                  <div>
                    <label className="block mb-2">DNI</label>
                    <div className="relative">
                      <IdCard className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                      <input
                        type="text"
                        value={formData.dni}
                        onChange={(e) => setFormData({ ...formData, dni: e.target.value })}
                        disabled={modalMode === 'view'}
                        className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                        placeholder="30111222"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block mb-2">Fecha de Nacimiento</label>
                    <div className="relative">
                      <Calendar className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                      <input
                        type="date"
                        value={formData.birthDate}
                        onChange={(e) => setFormData({ ...formData, birthDate: e.target.value })}
                        disabled={modalMode === 'view'}
                        className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                      />
                    </div>
                  </div>

                  <div className="md:col-span-2">
                    <label className="block mb-2">Dirección</label>
                    <div className="relative">
                      <MapPin className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                      <input
                        type="text"
                        value={formData.address}
                        onChange={(e) => setFormData({ ...formData, address: e.target.value })}
                        disabled={modalMode === 'view'}
                        className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                        placeholder="Calle 123, Alta Gracia"
                      />
                    </div>
                  </div>

                  {modalMode === 'create' && (
                    <div className="md:col-span-2">
                      <label className="block mb-2">Contraseña</label>
                      <div className="relative">
                        <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                        <input
                          type="password"
                          value={formData.password}
                          onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                          className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                          placeholder="••••••••"
                        />
                      </div>
                      <p className="text-xs text-muted-foreground mt-1">Mínimo 8 caracteres, una mayúscula y un carácter especial.</p>
                    </div>
                  )}
                </div>

                {modalMode !== 'view' && (
                  <div>
                    <label className="block mb-2">Foto de Perfil</label>
                    <div className="border-2 border-dashed border-border rounded-lg p-6 text-center hover:border-primary transition-colors cursor-pointer">
                      <Upload className="w-6 h-6 text-muted-foreground mx-auto mb-2" />
                      <p className="text-sm text-muted-foreground mb-1">Haz clic o arrastra una imagen aquí</p>
                      <p className="text-xs text-muted-foreground">JPG, JPEG, PNG o WEBP hasta 5MB</p>
                    </div>
                  </div>
                )}

                {modalMode !== 'view' && (
                  <div>
                    <label className="block mb-2">Documentación (DNI, certificados, etc.)</label>
                    <div className="border-2 border-dashed border-border rounded-lg p-6 text-center hover:border-primary transition-colors cursor-pointer">
                      <FileText className="w-6 h-6 text-muted-foreground mx-auto mb-2" />
                      <p className="text-sm text-muted-foreground mb-1">Haz clic o arrastra un archivo aquí</p>
                      <p className="text-xs text-muted-foreground">PDF, JPG, JPEG o PNG hasta 10MB por archivo</p>
                    </div>
                  </div>
                )}

                {formError && (
                  <div className="flex items-start gap-2 text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-lg p-3">
                    <AlertCircle className="w-4 h-4 flex-shrink-0 mt-0.5" />
                    <span>{formError}</span>
                  </div>
                )}

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
              <h2 className="text-xl font-semibold">Aprobar / Cambiar Estado</h2>
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
                  className="w-full px-4 py-2 rounded-lg border border-primary text-primary hover:bg-primary/5 transition-colors flex items-center justify-center gap-2"
                >
                  <CheckCircle className="w-4 h-4" />
                  Aprobar / Activar Usuario
                </button>
                <button
                  onClick={() => updateUserStatus('Pendiente')}
                  className="w-full px-4 py-2 rounded-lg border border-secondary text-secondary hover:bg-secondary/5 transition-colors flex items-center justify-center gap-2"
                >
                  <AlertCircle className="w-4 h-4" />
                  Rechazar / Poner en Pendiente
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

      {/* Delete Confirmation Modal */}
      {showDeleteModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-sm p-6">
            <h2 className="text-xl font-semibold mb-3">Eliminar Usuario</h2>
            <p className="text-sm text-muted-foreground mb-6">
              ¿Confirmás que querés eliminar a <strong className="text-foreground">{userToDelete?.name}</strong>?
              El perfil dejará de estar visible en las listas del sistema.
            </p>
            <div className="flex gap-3">
              <button
                onClick={() => setShowDeleteModal(false)}
                className="flex-1 px-4 py-2.5 rounded-lg border border-border hover:bg-secondary transition-colors"
              >
                Cancelar
              </button>
              <button
                onClick={confirmDeleteUser}
                className="flex-1 px-4 py-2.5 bg-destructive text-destructive-foreground rounded-lg hover:opacity-90 transition-opacity"
              >
                Eliminar
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
