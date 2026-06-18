import { useState } from 'react';
import { Link } from 'react-router-dom';
import { User, Mail, Lock, Phone, Leaf, Eye, EyeOff } from 'lucide-react';
import { InputText } from 'primereact/inputtext';
import { Password } from 'primereact/password';
import { Dropdown } from 'primereact/dropdown';
import { Button } from 'primereact/button';

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
            <div className="w-24 h-24 mx-auto rounded-full overflow-hidden bg-primary flex items-center justify-center mb-4 shadow-md">
              <img src="/logo.png" alt="Buen Vivir" className="w-full h-full object-cover" />
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
                  <User className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground z-10" />
                  <InputText
                    name="name"
                    value={formData.name}
                    onChange={handleChange as any}
                    className="w-full pl-10 pr-4 py-3 border border-border rounded-md"
                    placeholder="Juan Pérez"
                    required
                  />
                </div>
              </div>

              <div>
                <label className="block mb-2">Email</label>
                <div className="relative">
                  <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground z-10" />
                  <InputText
                    type="email"
                    name="email"
                    value={formData.email}
                    onChange={handleChange as any}
                    className="w-full pl-10 pr-4 py-3 border border-border rounded-md"
                    placeholder="tu@email.com"
                    required
                  />
                </div>
              </div>

              <div>
                <label className="block mb-2">Teléfono</label>
                <div className="relative">
                  <Phone className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground z-10" />
                  <InputText
                    type="tel"
                    name="phone"
                    value={formData.phone}
                    onChange={handleChange as any}
                    className="w-full pl-10 pr-4 py-3 border border-border rounded-md"
                    placeholder="+54 9 11 1234-5678"
                    required
                  />
                </div>
              </div>

              <div>
                <label className="block mb-2">Rol</label>
                <Dropdown
                  name="role"
                  value={formData.role}
                  options={[
                    { label: 'Participante', value: 'participante' },
                    { label: 'Voluntario', value: 'voluntario' },
                    { label: 'Docente', value: 'docente' },
                    { label: 'Empresa', value: 'empresa' }
                  ]}
                  onChange={(e) => setFormData({ ...formData, role: e.value })}
                  className="w-full border border-border rounded-md flex items-center"
                  pt={{ input: { className: 'py-3 px-4' } }}
                />
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
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground z-10" />
                  <Password
                    name="password"
                    value={formData.password}
                    onChange={handleChange as any}
                    toggleMask
                    className="w-full"
                    inputClassName="w-full pl-10 pr-12 py-3 border border-border rounded-md bg-transparent"
                    placeholder="••••••••"
                    required
                    feedback={false}
                  />
                </div>
              </div>

              <div>
                <label className="block mb-2">Confirmar Contraseña</label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground z-10" />
                  <Password
                    name="confirmPassword"
                    value={formData.confirmPassword}
                    onChange={handleChange as any}
                    toggleMask
                    className="w-full"
                    inputClassName="w-full pl-10 pr-12 py-3 border border-border rounded-md bg-transparent"
                    placeholder="••••••••"
                    required
                    feedback={false}
                  />
                </div>
              </div>

              <div className="flex items-start gap-2">
                <input type="checkbox" id="terms" className="mt-1 rounded border-border" required />
                <label htmlFor="terms" className="text-sm text-muted-foreground">
                  Acepto los términos y condiciones y la política de privacidad
                </label>
              </div>

              <Button
                type="submit"
                label="Crear Cuenta"
                className="w-full py-3 bg-primary text-primary-foreground hover:opacity-90 transition-opacity border-none"
              />
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
