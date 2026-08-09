import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Settings, LogOut, BookOpen, Briefcase, Megaphone, ShieldCheck, Mail, Smartphone, Lock, Link2Off, X, AlertCircle, UserCircle } from 'lucide-react';
import { Toast, ToastState } from './base';

const PASSWORD_REGEX = /^(?=.*[A-Z])(?=.*[^A-Za-z0-9]).{8,}$/;

interface NotificationSettingsProps {
  onLogout: () => void;
}

interface PreferenceRow {
  key: string;
  label: string;
  description: string;
  icon: typeof BookOpen;
}

const rows: PreferenceRow[] = [
  { key: 'courses', label: 'Cursos', description: 'Inscripciones, recordatorios de clase y avances', icon: BookOpen },
  { key: 'jobs', label: 'Ofertas Laborales', description: 'Nuevas ofertas compatibles con tu perfil', icon: Briefcase },
  { key: 'campaigns', label: 'Campañas', description: 'Novedades de campañas y concientización', icon: Megaphone },
  { key: 'system', label: 'Sistema', description: 'Aprobaciones de cuenta, moderación y seguridad', icon: ShieldCheck },
];

type Prefs = Record<string, { email: boolean; push: boolean }>;

const initialPrefs: Prefs = {
  courses: { email: true, push: true },
  jobs: { email: true, push: false },
  campaigns: { email: false, push: true },
  system: { email: true, push: true },
};

const emptyPasswordForm = { current: '', next: '', confirm: '' };

export function NotificationSettings({ onLogout }: NotificationSettingsProps) {
  const [prefs, setPrefs] = useState<Prefs>(initialPrefs);
  const [toast, setToast] = useState<ToastState | null>(null);
  const [googleLinked, setGoogleLinked] = useState(true);
  const [showPasswordModal, setShowPasswordModal] = useState(false);
  const [showUnlinkModal, setShowUnlinkModal] = useState(false);
  const [passwordForm, setPasswordForm] = useState(emptyPasswordForm);
  const [passwordError, setPasswordError] = useState('');

  const toggle = (key: string, channel: 'email' | 'push') => {
    setPrefs((prev) => ({
      ...prev,
      [key]: { ...prev[key], [channel]: !prev[key][channel] },
    }));
  };

  const handleSave = () => {
    setToast({ type: 'success', message: 'Preferencias de notificaciones guardadas correctamente.' });
  };

  const handleChangePassword = (e: React.FormEvent) => {
    e.preventDefault();
    const { current, next, confirm } = passwordForm;
    if (!current.trim() || !next.trim() || !confirm.trim()) {
      setPasswordError('Completá todos los campos.');
      return;
    }
    if (!PASSWORD_REGEX.test(next)) {
      setPasswordError('La nueva contraseña debe tener al menos 8 caracteres, una mayúscula y un carácter especial.');
      return;
    }
    if (next !== confirm) {
      setPasswordError('La confirmación no coincide con la nueva contraseña.');
      return;
    }
    setPasswordError('');
    setShowPasswordModal(false);
    setPasswordForm(emptyPasswordForm);
    setToast({ type: 'success', message: 'Contraseña actualizada correctamente.' });
  };

  const handleUnlinkGoogle = () => {
    setGoogleLinked(false);
    setShowUnlinkModal(false);
    setToast({ type: 'success', message: 'Tu cuenta de Google fue desvinculada.' });
  };

  return (
    <div className="min-h-screen bg-background">
      <Toast toast={toast} onClose={() => setToast(null)} />

      <header className="bg-card border-b border-border p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link to="/dashboard" className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center">
              <img src="/logo.png" alt="Buen Vivir" className="w-8 h-8 object-contain" />
            </Link>
            <div>
              <h1 className="text-2xl flex items-center gap-2">
                <Settings className="w-6 h-6 text-primary" />
                Configuración de Notificaciones
              </h1>
              <p className="text-sm text-muted-foreground">Elegí qué avisos querés recibir y por qué medio</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Link to="/notificaciones" className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors">
              Volver a Notificaciones
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

      <main className="p-6 max-w-3xl mx-auto">
        {/* Account Section */}
        <div className="bg-card rounded-lg border border-border overflow-hidden mb-6">
          <div className="px-6 py-3 bg-muted text-xs font-semibold text-muted-foreground uppercase tracking-wide flex items-center gap-2">
            <UserCircle className="w-3.5 h-3.5" />
            Cuenta
          </div>

          <div className="flex items-center justify-between gap-4 px-6 py-4 border-t border-border">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center flex-shrink-0">
                <Lock className="w-5 h-5 text-primary" />
              </div>
              <div>
                <p className="font-medium">Contraseña</p>
                <p className="text-xs text-muted-foreground">Actualizá la contraseña de tu cuenta</p>
              </div>
            </div>
            <button
              onClick={() => { setPasswordForm(emptyPasswordForm); setPasswordError(''); setShowPasswordModal(true); }}
              className="px-4 py-2 rounded-lg border border-border hover:bg-secondary transition-colors text-sm whitespace-nowrap"
            >
              Cambiar Contraseña
            </button>
          </div>

          <div className="flex items-center justify-between gap-4 px-6 py-4 border-t border-border">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center flex-shrink-0">
                <Link2Off className="w-5 h-5 text-primary" />
              </div>
              <div>
                <p className="font-medium">Cuenta de Google</p>
                <p className="text-xs text-muted-foreground">
                  {googleLinked ? 'Tu cuenta está vinculada con Google' : 'Tu cuenta ya no está vinculada con Google'}
                </p>
              </div>
            </div>
            <button
              onClick={() => setShowUnlinkModal(true)}
              disabled={!googleLinked}
              className="px-4 py-2 rounded-lg border border-destructive/30 text-destructive hover:bg-destructive/10 transition-colors text-sm whitespace-nowrap disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {googleLinked ? 'Desvincular cuenta de Google' : 'Desvinculada'}
            </button>
          </div>
        </div>

        <div className="bg-card rounded-lg border border-border overflow-hidden">
          <div className="grid grid-cols-[1fr_auto_auto] gap-4 items-center px-6 py-3 bg-muted text-xs font-semibold text-muted-foreground uppercase tracking-wide">
            <span>Tipo de notificación</span>
            <span className="flex items-center gap-1 justify-center"><Mail className="w-3.5 h-3.5" /> Email</span>
            <span className="flex items-center gap-1 justify-center"><Smartphone className="w-3.5 h-3.5" /> Push</span>
          </div>

          {rows.map((row) => (
            <div key={row.key} className="grid grid-cols-[1fr_auto_auto] gap-4 items-center px-6 py-4 border-t border-border">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center flex-shrink-0">
                  <row.icon className="w-5 h-5 text-primary" />
                </div>
                <div>
                  <p className="font-medium">{row.label}</p>
                  <p className="text-xs text-muted-foreground">{row.description}</p>
                </div>
              </div>

              <button
                onClick={() => toggle(row.key, 'email')}
                className={`w-11 h-6 rounded-full transition-colors relative flex-shrink-0 justify-self-center ${
                  prefs[row.key].email ? 'bg-primary' : 'bg-muted'
                }`}
                aria-label={`Notificaciones por email de ${row.label}`}
              >
                <span className={`absolute top-0.5 w-5 h-5 rounded-full bg-white shadow transition-transform ${
                  prefs[row.key].email ? 'translate-x-5' : 'translate-x-0.5'
                }`} />
              </button>

              <button
                onClick={() => toggle(row.key, 'push')}
                className={`w-11 h-6 rounded-full transition-colors relative flex-shrink-0 justify-self-center ${
                  prefs[row.key].push ? 'bg-primary' : 'bg-muted'
                }`}
                aria-label={`Notificaciones push de ${row.label}`}
              >
                <span className={`absolute top-0.5 w-5 h-5 rounded-full bg-white shadow transition-transform ${
                  prefs[row.key].push ? 'translate-x-5' : 'translate-x-0.5'
                }`} />
              </button>
            </div>
          ))}
        </div>

        <p className="text-xs text-muted-foreground mt-4">
          Las notificaciones push requieren tener la app instalada (PWA) y haber otorgado el permiso correspondiente.
        </p>

        <div className="flex justify-end mt-6">
          <button
            onClick={handleSave}
            className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
          >
            Guardar Preferencias
          </button>
        </div>
      </main>

      {/* Change Password Modal */}
      {showPasswordModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-md">
            <div className="border-b border-border p-6 flex items-center justify-between">
              <h2 className="text-xl font-semibold flex items-center gap-2">
                <Lock className="w-5 h-5 text-primary" />
                Cambiar Contraseña
              </h2>
              <button
                onClick={() => setShowPasswordModal(false)}
                className="p-2 rounded-lg hover:bg-secondary transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6">
              <form onSubmit={handleChangePassword} className="space-y-4">
                <div>
                  <label className="block mb-2 text-sm">Contraseña Actual</label>
                  <input
                    type="password"
                    value={passwordForm.current}
                    onChange={(e) => setPasswordForm({ ...passwordForm, current: e.target.value })}
                    className="w-full px-4 py-2.5 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="••••••••"
                  />
                </div>
                <div>
                  <label className="block mb-2 text-sm">Nueva Contraseña</label>
                  <input
                    type="password"
                    value={passwordForm.next}
                    onChange={(e) => setPasswordForm({ ...passwordForm, next: e.target.value })}
                    className="w-full px-4 py-2.5 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="••••••••"
                  />
                  <p className="text-xs text-muted-foreground mt-1">Mínimo 8 caracteres, una mayúscula y un carácter especial.</p>
                </div>
                <div>
                  <label className="block mb-2 text-sm">Confirmar Nueva Contraseña</label>
                  <input
                    type="password"
                    value={passwordForm.confirm}
                    onChange={(e) => setPasswordForm({ ...passwordForm, confirm: e.target.value })}
                    className="w-full px-4 py-2.5 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="••••••••"
                  />
                </div>

                {passwordError && (
                  <div className="flex items-start gap-2 text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-lg p-3">
                    <AlertCircle className="w-4 h-4 flex-shrink-0 mt-0.5" />
                    <span>{passwordError}</span>
                  </div>
                )}

                <div className="flex gap-3 pt-2">
                  <button
                    type="button"
                    onClick={() => setShowPasswordModal(false)}
                    className="flex-1 px-4 py-2.5 rounded-lg border border-border hover:bg-secondary transition-colors text-sm"
                  >
                    Cancelar
                  </button>
                  <button
                    type="submit"
                    className="flex-1 px-4 py-2.5 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity text-sm"
                  >
                    Guardar
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Unlink Google Modal */}
      {showUnlinkModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg border border-border w-full max-w-sm p-6">
            <h2 className="text-xl font-semibold mb-3">Desvincular cuenta de Google</h2>
            <p className="text-sm text-muted-foreground mb-6">
              ¿Confirmás que querés desvincular tu cuenta de Google? Vas a poder seguir ingresando con tu email y contraseña.
            </p>
            <div className="flex gap-3">
              <button
                onClick={() => setShowUnlinkModal(false)}
                className="flex-1 px-4 py-2.5 rounded-lg border border-border hover:bg-secondary transition-colors"
              >
                Cancelar
              </button>
              <button
                onClick={handleUnlinkGoogle}
                className="flex-1 px-4 py-2.5 bg-destructive text-destructive-foreground rounded-lg hover:opacity-90 transition-opacity"
              >
                Desvincular
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
