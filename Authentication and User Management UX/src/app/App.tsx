import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useState } from 'react';
import { LandingPage } from './components/LandingPage';
import { Login } from './components/Login';
import { ForgotPassword } from './components/ForgotPassword';
import { Register } from './components/Register';
import { Dashboard } from './components/Dashboard';
import { UserManagement } from './components/UserManagement';
import { AuditLog } from './components/AuditLog';
import { VolunteerUserManagement } from './components/VolunteerUserManagement';
import { AttendanceTracking } from './components/AttendanceTracking';
import { CourseCatalog } from './components/CourseCatalog';
import { CourseDetail } from './components/CourseDetail';
import { ProgressTracking } from './components/ProgressTracking';
import { TeacherPortal } from './components/TeacherPortal';
import { PublicJobOffers } from './components/PublicJobOffers';
import { JobOffers } from './components/JobOffers';
import { PublicCampaigns } from './components/PublicCampaigns';
import { ParticipantTracking } from './components/ParticipantTracking';
import { CVGenerator } from './components/CVGenerator';
import { CampaignManagement } from './components/CampaignManagement';
import { LaborInsertion } from './components/LaborInsertion';
import { NotificationCenter } from './components/NotificationCenter';
import { NotificationSettings } from './components/NotificationSettings';

export default function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [userRole, setUserRole] = useState<'admin' | 'participante' | 'voluntario' | 'empresa' | 'docente'>('participante');

  const handleLogin = (role: typeof userRole) => {
    setIsAuthenticated(true);
    setUserRole(role);
  };

  const handleLogout = () => {
    setIsAuthenticated(false);
  };

  return (
    <BrowserRouter>
      <Routes>
        {/* Public Routes */}
        <Route path="/" element={<LandingPage />} />
        <Route path="/ofertas-laborales-publico" element={<PublicJobOffers />} />
        <Route path="/campanas-publico" element={<PublicCampaigns />} />

        {/* Auth Routes */}
        <Route
          path="/login"
          element={
            isAuthenticated ?
            <Navigate to="/dashboard" /> :
            <Login onLogin={handleLogin} />
          }
        />
        <Route
          path="/register"
          element={
            isAuthenticated ?
            <Navigate to="/dashboard" /> :
            <Register />
          }
        />
        <Route path="/recuperar-password" element={<ForgotPassword />} />

        {/* Protected Routes */}
        <Route
          path="/dashboard"
          element={
            isAuthenticated ?
            <Dashboard userRole={userRole} onLogout={handleLogout} /> :
            <Navigate to="/login" />
          }
        />

        {/* Admin Routes */}
        <Route
          path="/users"
          element={
            isAuthenticated && userRole === 'admin' ?
            <UserManagement onLogout={handleLogout} /> :
            <Navigate to="/dashboard" />
          }
        />
        <Route
          path="/log-auditoria"
          element={
            isAuthenticated && userRole === 'admin' ?
            <AuditLog onLogout={handleLogout} /> :
            <Navigate to="/dashboard" />
          }
        />
        {/* Volunteer Routes */}
        <Route
          path="/seguimiento-participantes"
          element={
            isAuthenticated && (userRole === 'admin' || userRole === 'voluntario') ?
            <ParticipantTracking userRole={userRole} onLogout={handleLogout} /> :
            <Navigate to="/dashboard" />
          }
        />
        <Route
          path="/asistencia-general"
          element={
            isAuthenticated && (userRole === 'admin' || userRole === 'voluntario') ?
            <AttendanceTracking onLogout={handleLogout} /> :
            <Navigate to="/dashboard" />
          }
        />

        {/* Campaign Management - Admin & Volunteer */}
        <Route
          path="/gestion-campanas"
          element={
            isAuthenticated && (userRole === 'admin' || userRole === 'voluntario') ?
            <CampaignManagement userRole={userRole} onLogout={handleLogout} /> :
            <Navigate to="/dashboard" />
          }
        />

        {/* Participant Routes */}
        <Route
          path="/generar-cv"
          element={
            isAuthenticated && userRole === 'participante' ?
            <CVGenerator onLogout={handleLogout} /> :
            <Navigate to="/dashboard" />
          }
        />
        <Route
          path="/ofertas-laborales"
          element={
            isAuthenticated && userRole === 'participante' ?
            <JobOffers onLogout={handleLogout} /> :
            <Navigate to="/dashboard" />
          }
        />

        {/* Course Routes */}
        <Route
          path="/courses"
          element={
            isAuthenticated ?
            <CourseCatalog userRole={userRole} onLogout={handleLogout} /> :
            <Navigate to="/login" />
          }
        />
        <Route
          path="/courses/:id"
          element={
            isAuthenticated ?
            <CourseDetail userRole={userRole} onLogout={handleLogout} /> :
            <Navigate to="/login" />
          }
        />
        <Route
          path="/progress"
          element={
            isAuthenticated ?
            <ProgressTracking userRole={userRole} onLogout={handleLogout} /> :
            <Navigate to="/login" />
          }
        />

        {/* Teacher Portal */}
        <Route
          path="/teacher-portal"
          element={
            isAuthenticated && userRole === 'docente' ?
            <TeacherPortal onLogout={handleLogout} /> :
            <Navigate to="/dashboard" />
          }
        />

        {/* Labor Insertion */}
        <Route
          path="/insercion-laboral"
          element={
            isAuthenticated && (userRole === 'admin' || userRole === 'voluntario' || userRole === 'empresa') ?
            <LaborInsertion userRole={userRole} onLogout={handleLogout} /> :
            <Navigate to="/dashboard" />
          }
        />

        {/* Volunteer User Management */}
        <Route
          path="/gestion-participantes"
          element={
            isAuthenticated && userRole === 'voluntario' ?
            <VolunteerUserManagement onLogout={handleLogout} /> :
            <Navigate to="/dashboard" />
          }
        />

        {/* Notifications - any authenticated role */}
        <Route
          path="/notificaciones"
          element={
            isAuthenticated ?
            <NotificationCenter onLogout={handleLogout} /> :
            <Navigate to="/login" />
          }
        />
        <Route
          path="/configuracion-notificaciones"
          element={
            isAuthenticated ?
            <NotificationSettings onLogout={handleLogout} /> :
            <Navigate to="/login" />
          }
        />
      </Routes>
    </BrowserRouter>
  );
}
