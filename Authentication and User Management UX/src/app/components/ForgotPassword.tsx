import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Mail, Lock, ArrowLeft, CheckCircle2 } from 'lucide-react';
import { InputText } from 'primereact/inputtext';
import { Password } from 'primereact/password';
import { Button } from 'primereact/button';

type Step = 'request' | 'sent' | 'reset' | 'done';

export function ForgotPassword() {
  const [step, setStep] = useState<Step>('request');
  const [email, setEmail] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');

  const handleRequestSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    // Mensaje siempre genérico: no confirma ni desmiente si el email existe (HU-02bis).
    setStep('sent');
  };

  const passwordPolicyOk = (pwd: string) =>
    pwd.length >= 8 && /[A-Z]/.test(pwd) && /[^A-Za-z0-9]/.test(pwd);

  const handleResetSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!passwordPolicyOk(newPassword)) {
      setError('La contraseña debe tener al menos 8 caracteres, una mayúscula y un carácter especial.');
      return;
    }
    if (newPassword !== confirmPassword) {
      setError('Las contraseñas no coinciden.');
      return;
    }
    setError('');
    setStep('done');
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-8 bg-background">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="w-24 h-24 mx-auto rounded-full overflow-hidden bg-primary flex items-center justify-center mb-4 shadow-md">
            <img src="/logo.png" alt="Buen Vivir" className="w-full h-full object-cover" />
          </div>
          <h1 className="text-3xl mb-2">Recuperar Contraseña</h1>
          <p className="text-muted-foreground">Centro de Ecología Integral</p>
        </div>

        <div className="bg-card rounded-lg border border-border p-8 shadow-sm">
          {step === 'request' && (
            <form onSubmit={handleRequestSubmit} className="space-y-4">
              <p className="text-sm text-muted-foreground">
                Ingresá tu email y, si está registrado, te enviaremos un enlace para restablecer tu contraseña.
              </p>
              <div>
                <label className="block mb-2">Email</label>
                <div className="relative">
                  <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground z-10" />
                  <InputText
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full pl-10 pr-4 py-3 border border-border rounded-md"
                    placeholder="tu@email.com"
                    required
                  />
                </div>
              </div>
              <Button
                type="submit"
                label="Enviar enlace de recuperación"
                className="w-full py-3 bg-primary text-primary-foreground hover:opacity-90 transition-opacity border-none"
              />
            </form>
          )}

          {step === 'sent' && (
            <div className="text-center py-4">
              <div className="w-14 h-14 rounded-full bg-primary/10 flex items-center justify-center mx-auto mb-4">
                <Mail className="w-7 h-7 text-primary" />
              </div>
              <h3 className="text-lg mb-2">Revisá tu correo</h3>
              <p className="text-sm text-muted-foreground mb-6">
                Si el email <strong className="text-foreground">{email}</strong> está registrado, vas a recibir un enlace
                de recuperación válido por 30 minutos y de un solo uso.
              </p>
              <button
                onClick={() => setStep('reset')}
                className="text-sm text-primary hover:underline"
              >
                (Demo) Simular clic en el enlace del email
              </button>
            </div>
          )}

          {step === 'reset' && (
            <form onSubmit={handleResetSubmit} className="space-y-4">
              <p className="text-sm text-muted-foreground">Ingresá tu nueva contraseña.</p>
              <div>
                <label className="block mb-2">Nueva contraseña</label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground z-10" />
                  <Password
                    value={newPassword}
                    onChange={(e: any) => setNewPassword(e.target.value)}
                    toggleMask
                    className="w-full"
                    inputClassName="w-full pl-10 pr-12 py-3 border border-border rounded-md bg-transparent"
                    placeholder="••••••••"
                    feedback={false}
                    required
                  />
                </div>
              </div>
              <div>
                <label className="block mb-2">Confirmar contraseña</label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground z-10" />
                  <Password
                    value={confirmPassword}
                    onChange={(e: any) => setConfirmPassword(e.target.value)}
                    toggleMask
                    className="w-full"
                    inputClassName="w-full pl-10 pr-12 py-3 border border-border rounded-md bg-transparent"
                    placeholder="••••••••"
                    feedback={false}
                    required
                  />
                </div>
              </div>
              <p className="text-xs text-muted-foreground">
                Mínimo 8 caracteres, una mayúscula y un carácter especial.
              </p>
              {error && <p className="text-sm text-destructive">{error}</p>}
              <Button
                type="submit"
                label="Restablecer contraseña"
                className="w-full py-3 bg-primary text-primary-foreground hover:opacity-90 transition-opacity border-none"
              />
            </form>
          )}

          {step === 'done' && (
            <div className="text-center py-4">
              <div className="w-14 h-14 rounded-full bg-primary/10 flex items-center justify-center mx-auto mb-4">
                <CheckCircle2 className="w-7 h-7 text-primary" />
              </div>
              <h3 className="text-lg mb-2">Contraseña actualizada</h3>
              <p className="text-sm text-muted-foreground mb-6">
                Tu contraseña se actualizó correctamente. Por seguridad, cerramos todas tus sesiones activas anteriores.
              </p>
              <Link
                to="/login"
                className="inline-block px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
              >
                Iniciar sesión
              </Link>
            </div>
          )}
        </div>

        <p className="text-center mt-6 text-sm text-muted-foreground">
          <Link to="/login" className="text-primary hover:underline inline-flex items-center gap-1">
            <ArrowLeft className="w-4 h-4" />
            Volver a Iniciar Sesión
          </Link>
        </p>
      </div>
    </div>
  );
}
