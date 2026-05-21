import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Mail, Lock, Eye, EyeOff, Leaf, Shield } from 'lucide-react';

interface LoginProps {
  onLogin: (role: 'admin' | 'participante' | 'voluntario' | 'empresa' | 'docente') => void;
}

export function Login({ onLogin }: LoginProps) {
  const [showPassword, setShowPassword] = useState(false);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [show2FA, setShow2FA] = useState(false);
  const [twoFactorCode, setTwoFactorCode] = useState('');
  const [selectedRole, setSelectedRole] = useState<'admin' | 'participante' | 'voluntario' | 'empresa' | 'docente'>('participante');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!show2FA) {
      setShow2FA(true);
    } else {
      onLogin(selectedRole);
    }
  };

  const handleGoogleLogin = () => {
    onLogin(selectedRole);
  };

  return (
    <div className="min-h-screen flex">
      {/* Left side - Login form */}
      <div className="flex-1 flex items-center justify-center p-8 bg-background">
        <div className="w-full max-w-md">
          {/* Logo and Title */}
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary mb-4">
              <Leaf className="w-8 h-8 text-primary-foreground" />
            </div>
            <h1 className="text-3xl mb-2">Buen Vivir</h1>
            <p className="text-muted-foreground">
              Centro de Ecología Integral
            </p>
          </div>

          {/* Login Form */}
          <div className="bg-card rounded-lg border border-border p-8 shadow-sm">
            <h2 className="text-2xl mb-6">Iniciar Sesión</h2>

            <form onSubmit={handleSubmit} className="space-y-4">
              {!show2FA ? (
                <>
                  <div>
                    <label className="block mb-2">Tipo de Usuario (Demo)</label>
                    <select
                      value={selectedRole}
                      onChange={(e) => setSelectedRole(e.target.value as any)}
                      className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    >
                      <option value="admin">Responsable (ve todo)</option>
                      <option value="voluntario">Voluntario (seguimiento y campañas)</option>
                      <option value="participante">Participante (cursos y CV)</option>
                      <option value="docente">Docente (subir contenido)</option>
                      <option value="empresa">Empresa</option>
                    </select>
                  </div>

                  <div>
                    <label className="block mb-2">Email o Usuario</label>
                    <div className="relative">
                      <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                      <input
                        type="text"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                        placeholder="tu@email.com"
                        required
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block mb-2">Contraseña</label>
                    <div className="relative">
                      <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                      <input
                        type={showPassword ? 'text' : 'password'}
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        className="w-full pl-10 pr-12 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                        placeholder="••••••••"
                        required
                      />
                      <button
                        type="button"
                        onClick={() => setShowPassword(!showPassword)}
                        className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                      >
                        {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                      </button>
                    </div>
                  </div>

                  <div className="flex items-center justify-between text-sm">
                    <label className="flex items-center gap-2">
                      <input type="checkbox" className="rounded border-border" />
                      <span>Recordarme</span>
                    </label>
                    <a href="#" className="text-primary hover:underline">
                      ¿Olvidaste tu contraseña?
                    </a>
                  </div>

                  <button
                    type="submit"
                    className="w-full py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
                  >
                    Continuar
                  </button>
                </>
              ) : (
                <>
                  <div className="text-center mb-4">
                    <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-secondary mb-3">
                      <Shield className="w-6 h-6 text-primary" />
                    </div>
                    <h3 className="text-xl mb-2">Autenticación de Doble Factor</h3>
                    <p className="text-sm text-muted-foreground">
                      Ingresa el código de 6 dígitos de tu aplicación de autenticación
                    </p>
                  </div>

                  <div>
                    <label className="block mb-2">Código 2FA</label>
                    <input
                      type="text"
                      value={twoFactorCode}
                      onChange={(e) => setTwoFactorCode(e.target.value)}
                      className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary text-center text-2xl tracking-widest"
                      placeholder="000000"
                      maxLength={6}
                      required
                    />
                  </div>

                  <button
                    type="submit"
                    className="w-full py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
                  >
                    Verificar e Ingresar
                  </button>

                  <button
                    type="button"
                    onClick={() => setShow2FA(false)}
                    className="w-full py-2 text-primary hover:underline"
                  >
                    Volver
                  </button>
                </>
              )}
            </form>

            {!show2FA && (
              <>
                <div className="relative my-6">
                  <div className="absolute inset-0 flex items-center">
                    <div className="w-full border-t border-border"></div>
                  </div>
                  <div className="relative flex justify-center text-sm">
                    <span className="px-2 bg-card text-muted-foreground">O continúa con</span>
                  </div>
                </div>

                <button
                  onClick={handleGoogleLogin}
                  className="w-full py-3 border border-border rounded-lg hover:bg-secondary transition-colors flex items-center justify-center gap-3"
                >
                  <svg className="w-5 h-5" viewBox="0 0 24 24">
                    <path
                      fill="#4285F4"
                      d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                    />
                    <path
                      fill="#34A853"
                      d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                    />
                    <path
                      fill="#FBBC05"
                      d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                    />
                    <path
                      fill="#EA4335"
                      d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                    />
                  </svg>
                  Google
                </button>
              </>
            )}
          </div>

          <p className="text-center mt-6 text-sm text-muted-foreground">
            ¿No tienes cuenta?{' '}
            <Link to="/register" className="text-primary hover:underline">
              Regístrate aquí
            </Link>
             <br>
            </br>
            <br>
            </br>
            
            <Link to="/" className="text-primary hover:underline">
              Volver a la pantalla principal
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
