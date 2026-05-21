import { useState } from 'react';
import { Link } from 'react-router-dom';
import { User, Mail, Lock, Phone, Leaf, Eye, EyeOff } from 'lucide-react';

export function Register() {
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    phone: '',
    password: '',
    confirmPassword: '',
    role: 'participante'
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    });
  };

  return (
    <div className="min-h-screen flex">
      {/* Left side - Registration form */}
      <div className="flex-1 flex items-center justify-center p-8 bg-background">
        <div className="w-full max-w-md">
          {/* Logo and Title */}
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary mb-4">
              <Leaf className="w-8 h-8 text-primary-foreground" />
            </div>
            <h1 className="text-3xl mb-2">Crear Cuenta</h1>
            <p className="text-muted-foreground">
              Únete a nuestra comunidad
            </p>
          </div>

          {/* Registration Form */}
          <div className="bg-card rounded-lg border border-border p-8 shadow-sm">
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block mb-2">Nombre Completo</label>
                <div className="relative">
                  <User className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                  <input
                    type="text"
                    name="name"
                    value={formData.name}
                    onChange={handleChange}
                    className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="Juan Pérez"
                    required
                  />
                </div>
              </div>

              <div>
                <label className="block mb-2">Email</label>
                <div className="relative">
                  <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                  <input
                    type="email"
                    name="email"
                    value={formData.email}
                    onChange={handleChange}
                    className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="tu@email.com"
                    required
                  />
                </div>
              </div>

              <div>
                <label className="block mb-2">Teléfono</label>
                <div className="relative">
                  <Phone className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                  <input
                    type="tel"
                    name="phone"
                    value={formData.phone}
                    onChange={handleChange}
                    className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="+54 9 11 1234-5678"
                    required
                  />
                </div>
              </div>

              <div>
                <label className="block mb-2">Rol</label>
                <select
                  name="role"
                  value={formData.role}
                  onChange={handleChange}
                  className="w-full px-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                  required
                >
                  <option value="participante">Participante</option>
                  <option value="voluntario">Voluntario</option>
                  <option value="docente">Docente</option>
                  <option value="empresa">Empresa</option>
                </select>
                <p className="text-xs text-muted-foreground mt-1">
                  {formData.role === 'voluntario' && 'Los voluntarios deben ser aprobados por los coordinadores'}
                  {formData.role === 'docente' && 'Los docentes tendrán acceso al portal educativo'}
                  {formData.role === 'participante' && 'Podrás acceder a cursos y talleres'}
                  {formData.role === 'empresa' && 'Colabora con oportunidades laborales'}
                </p>
              </div>

              <div>
                <label className="block mb-2">Contraseña</label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                  <input
                    type={showPassword ? 'text' : 'password'}
                    name="password"
                    value={formData.password}
                    onChange={handleChange}
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

              <div>
                <label className="block mb-2">Confirmar Contraseña</label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                  <input
                    type={showConfirmPassword ? 'text' : 'password'}
                    name="confirmPassword"
                    value={formData.confirmPassword}
                    onChange={handleChange}
                    className="w-full pl-10 pr-12 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="••••••••"
                    required
                  />
                  <button
                    type="button"
                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  >
                    {showConfirmPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                  </button>
                </div>
              </div>

              <div className="flex items-start gap-2">
                <input type="checkbox" id="terms" className="mt-1 rounded border-border" required />
                <label htmlFor="terms" className="text-sm text-muted-foreground">
                  Acepto los términos y condiciones y la política de privacidad
                </label>
              </div>

              <button
                type="submit"
                className="w-full py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
              >
                Crear Cuenta
              </button>
            </form>
          </div>

          <p className="text-center mt-6 text-sm text-muted-foreground">
            ¿Ya tienes cuenta?{' '}
            <Link to="/login" className="text-primary hover:underline">
              Inicia sesión
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
