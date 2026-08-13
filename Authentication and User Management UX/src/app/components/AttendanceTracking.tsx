import { useState } from 'react';
import { Link } from 'react-router-dom';
import { ArrowLeft, LogOut, Calendar, User, Check, X as XIcon } from 'lucide-react';
import { Toast, ToastState } from './base';

interface AttendanceTrackingProps {
  onLogout: () => void;
}

type AttendanceStatus = 'presente' | 'ausente';

interface AttendanceRecord {
  status: AttendanceStatus;
  observation: string;
}

const activeParticipants = [
  { id: 1, nombre: 'María', apellido: 'González' },
  { id: 2, nombre: 'Juan', apellido: 'Pérez' },
  { id: 3, nombre: 'Ana', apellido: 'Martínez' },
  { id: 4, nombre: 'Carlos', apellido: 'López' },
  { id: 5, nombre: 'Laura', apellido: 'Rodríguez' },
];

const today = new Date().toISOString().split('T')[0];

const recordKey = (date: string, participantId: number) => `${date}::${participantId}`;

export function AttendanceTracking({ onLogout }: AttendanceTrackingProps) {
  const [selectedDate, setSelectedDate] = useState(today);
  const [records, setRecords] = useState<Record<string, AttendanceRecord>>({});
  const [toast, setToast] = useState<ToastState | null>(null);

  const setStatus = (participantId: number, status: AttendanceStatus) => {
    const key = recordKey(selectedDate, participantId);
    setRecords((prev) => ({
      ...prev,
      [key]: { status, observation: prev[key]?.observation ?? '' },
    }));
  };

  const setObservation = (participantId: number, observation: string) => {
    const key = recordKey(selectedDate, participantId);
    setRecords((prev) => ({
      ...prev,
      [key]: { status: prev[key]?.status ?? 'presente', observation },
    }));
  };

  const markedCount = activeParticipants.filter((p) => records[recordKey(selectedDate, p.id)]).length;

  const handleSave = () => {
    setToast({
      type: 'success',
      message: `Asistencia del ${selectedDate} guardada correctamente (${markedCount}/${activeParticipants.length} participantes registrados).`,
    });
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
              <h1 className="text-2xl text-white flex items-center gap-2">
                Registro de Asistencia
              </h1>
              <p className="text-sm text-white/70">Control diario de concurrencia del centro</p>
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
      <main className="p-6 max-w-4xl mx-auto">
        <div className="bg-card rounded-lg border border-border p-6 mb-6 flex flex-col sm:flex-row items-start sm:items-center gap-4">
          <label className="flex items-center gap-2 text-sm font-medium">
            <Calendar className="w-4 h-4 text-primary" />
            Fecha
          </label>
          <input
            type="date"
            value={selectedDate}
            onChange={(e) => setSelectedDate(e.target.value)}
            className="px-4 py-2 rounded-lg border border-border bg-input-background focus:outline-none focus:ring-2 focus:ring-primary"
          />
          <span className="text-sm text-muted-foreground sm:ml-auto">
            {markedCount}/{activeParticipants.length} participantes registrados para esta fecha
          </span>
        </div>

        <div className="bg-card rounded-lg border border-border overflow-hidden">
          <div className="divide-y divide-border">
            {activeParticipants.map((participant) => {
              const record = records[recordKey(selectedDate, participant.id)];
              return (
                <div key={participant.id} className="p-4 flex flex-col sm:flex-row sm:items-center gap-3">
                  <div className="flex items-center gap-3 sm:w-56 flex-shrink-0">
                    <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                      <User className="w-5 h-5 text-primary" />
                    </div>
                    <span className="font-medium">{participant.nombre} {participant.apellido}</span>
                  </div>

                  <div className="flex gap-2 flex-shrink-0">
                    <button
                      onClick={() => setStatus(participant.id, 'presente')}
                      className={`px-3 py-2 rounded-lg text-sm font-medium flex items-center gap-1.5 transition-colors ${
                        record?.status === 'presente'
                          ? 'bg-primary text-primary-foreground'
                          : 'border border-border hover:bg-secondary'
                      }`}
                    >
                      <Check className="w-4 h-4" />
                      Presente
                    </button>
                    <button
                      onClick={() => setStatus(participant.id, 'ausente')}
                      className={`px-3 py-2 rounded-lg text-sm font-medium flex items-center gap-1.5 transition-colors ${
                        record?.status === 'ausente'
                          ? 'bg-destructive text-destructive-foreground'
                          : 'border border-border hover:bg-secondary'
                      }`}
                    >
                      <XIcon className="w-4 h-4" />
                      Ausente
                    </button>
                  </div>

                  <input
                    type="text"
                    value={record?.observation ?? ''}
                    onChange={(e) => setObservation(participant.id, e.target.value)}
                    placeholder="Observación (opcional)"
                    className="flex-1 px-3 py-2 rounded-lg border border-border bg-input-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>
              );
            })}
          </div>
        </div>
        <div className="flex justify-end mt-6">
          <button
            onClick={handleSave}
            className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
          >
            Guardar Asistencia
          </button>
        </div>
      </main>
    </div>
  );
}
