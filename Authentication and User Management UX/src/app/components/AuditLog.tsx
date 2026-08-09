import { useState } from 'react';
import { Link } from 'react-router-dom';
import { LogOut, Search, ShieldCheck, Filter, ArrowLeft } from 'lucide-react';

interface AuditLogProps {
  onLogout: () => void;
}

const actionStyles: Record<string, string> = {
  Creación: 'bg-primary/10 text-primary',
  Modificación: 'bg-secondary/10 text-secondary',
  Eliminación: 'bg-destructive/10 text-destructive',
  Acceso: 'bg-[#1D1D1B]/10 text-[#1D1D1B]',
};

const entries = [
  { id: 1, user: 'admin@buenvivir.org', action: 'Modificación', resource: 'Participante: María González', date: '2026-08-07 09:12' },
  { id: 2, user: 'admin@buenvivir.org', action: 'Creación', resource: 'Usuario: Laura Rodríguez', date: '2026-08-06 16:40' },
  { id: 3, user: 'voluntario1@buenvivir.org', action: 'Creación', resource: 'Hito: Juan Pérez - Primera entrevista laboral', date: '2026-08-06 11:05' },
  { id: 4, user: 'admin@buenvivir.org', action: 'Eliminación', resource: 'Oferta laboral: Atención al Cliente', date: '2026-08-05 18:22' },
  { id: 5, user: 'desconocido', action: 'Acceso', resource: 'Intento de acceso denegado a /log-auditoria (403)', date: '2026-08-05 14:03' },
  { id: 6, user: 'admin@buenvivir.org', action: 'Modificación', resource: 'Curso: Panadería Artesanal', date: '2026-08-04 10:47' },
];

const actions = ['Todas', 'Creación', 'Modificación', 'Eliminación', 'Acceso'];

export function AuditLog({ onLogout }: AuditLogProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [actionFilter, setActionFilter] = useState('Todas');

  const filtered = entries.filter(entry => {
    const matchesSearch = entry.user.toLowerCase().includes(searchQuery.toLowerCase()) ||
      entry.resource.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesAction = actionFilter === 'Todas' || entry.action === actionFilter;
    return matchesSearch && matchesAction;
  });

  return (
    <div className="min-h-screen bg-background">
      <header className="bg-card border-b border-border p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link to="/dashboard" className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center">
              <img src="/logo.png" alt="Buen Vivir" className="w-8 h-8 object-contain" />
            </Link>
            <div>
              <h1 className="text-2xl flex items-center gap-2">
                <ShieldCheck className="w-6 h-6 text-primary" />
                Log de Auditoría
              </h1>
              <p className="text-sm text-muted-foreground">Registro inmutable de acciones críticas del sistema</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
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

      <main className="p-6">
        <div className="bg-card rounded-lg border border-border p-6 mb-6">
          <div className="flex flex-col lg:flex-row gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="Buscar por usuario o recurso..."
              />
            </div>
            <div className="relative">
              <Filter className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <select
                value={actionFilter}
                onChange={(e) => setActionFilter(e.target.value)}
                className="pl-9 pr-4 py-3 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
              >
                {actions.map((a) => <option key={a} value={a}>{a}</option>)}
              </select>
            </div>
          </div>
        </div>

        <div className="bg-card rounded-lg border border-border overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-muted">
                <tr>
                  <th className="text-left p-4">Usuario</th>
                  <th className="text-left p-4">Acción</th>
                  <th className="text-left p-4">Recurso</th>
                  <th className="text-left p-4">Fecha y hora</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((entry) => (
                  <tr key={entry.id} className="border-t border-border hover:bg-muted/50 transition-colors">
                    <td className="p-4 text-muted-foreground">{entry.user}</td>
                    <td className="p-4">
                      <span className={`inline-flex px-3 py-1 rounded-full text-xs font-medium ${actionStyles[entry.action]}`}>
                        {entry.action}
                      </span>
                    </td>
                    <td className="p-4">{entry.resource}</td>
                    <td className="p-4 text-muted-foreground whitespace-nowrap">{entry.date}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {filtered.length === 0 && (
          <div className="text-center py-12">
            <p className="text-muted-foreground text-lg">No se encontraron registros</p>
          </div>
        )}

        <p className="text-xs text-muted-foreground mt-4">
          Este log es de solo lectura: ninguna acción puede editarse ni eliminarse una vez registrada.
        </p>
      </main>
    </div>
  );
}
