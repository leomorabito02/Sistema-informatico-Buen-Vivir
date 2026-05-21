import React from 'react';
import { Card, CardHeader, CardBody } from './base';
import { CheckCircle, Clock, Target, Award, Plus } from 'lucide-react';

interface Milestone {
  id: number;
  title: string;
  description: string;
  date: string;
  status: 'completed' | 'in-progress' | 'pending';
  category: string;
}

interface PersonalPromotionProps {
  participantName: string;
  userId: string;
}

const mockMilestones: Milestone[] = [
  {
    id: 1,
    title: 'Completó Panadería Básica',
    description: 'Finalizó el curso de panadería artesanal con calificación 9.2/10',
    date: '2026-04-15',
    status: 'completed',
    category: 'Educación',
  },
  {
    id: 2,
    title: 'Primera Entrevista Laboral',
    description: 'Participó en entrevista para posición en Pastelería Local',
    date: '2026-05-01',
    status: 'completed',
    category: 'Empleabilidad',
  },
  {
    id: 3,
    title: 'Obtiene Certificado',
    description: 'Certificado de capacitación en Gastronomía y Panadería',
    date: '2026-05-20',
    status: 'completed',
    category: 'Logro',
  },
  {
    id: 4,
    title: 'En progreso: Costura y Diseño',
    description: 'Cursando Técnicas de Costura - Módulo 5 completado',
    date: '2026-05-21',
    status: 'in-progress',
    category: 'Educación',
  },
  {
    id: 5,
    title: 'Próxima: Entrevista Empresa XYZ',
    description: 'Entrevista programada para posición de especialista',
    date: '2026-06-10',
    status: 'pending',
    category: 'Empleabilidad',
  },
];

const keyIndicators = [
  { label: 'Cursos Completados', value: '3', icon: Award, color: 'bg-green-100 text-green-700' },
  { label: 'Cursos En Progreso', value: '2', icon: Clock, color: 'bg-blue-100 text-blue-700' },
  { label: 'Próximos Hitos', value: '4', icon: Target, color: 'bg-purple-100 text-purple-700' },
  { label: 'Certificados', value: '2', icon: Award, color: 'bg-yellow-100 text-yellow-700' },
];

export function PersonalPromotion({ participantName, userId }: PersonalPromotionProps) {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg p-6 text-white">
        <h1 className="text-3xl font-bold mb-2">Proceso de Promoción Personal</h1>
        <p className="text-blue-100">Seguimiento de hitos y avances de {participantName}</p>
      </div>

      {/* Key Indicators */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {keyIndicators.map((indicator) => (
          <Card key={indicator.label}>
            <CardBody>
              <div className="flex items-center gap-3">
                <div className={`w-12 h-12 rounded-lg flex items-center justify-center ${indicator.color}`}>
                  <indicator.icon size={24} />
                </div>
                <div>
                  <p className="text-2xl font-bold text-gray-900">{indicator.value}</p>
                  <p className="text-xs text-gray-600">{indicator.label}</p>
                </div>
              </div>
            </CardBody>
          </Card>
        ))}
      </div>

      {/* Timeline */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold flex items-center gap-2">
              <Target size={24} className="text-blue-600" />
              Línea de Tiempo de Hitos
            </h2>
            <button className="px-3 py-1 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700 transition-colors flex items-center gap-2">
              <Plus size={16} />
              Agregar Hito
            </button>
          </div>
        </CardHeader>
        <CardBody>
          <div className="relative">
            {/* Timeline Line */}
            <div className="absolute left-6 top-0 bottom-0 w-0.5 bg-gray-200"></div>

            {/* Timeline Items */}
            <div className="space-y-8">
              {mockMilestones.map((milestone, index) => (
                <div key={milestone.id} className="relative pl-20">
                  {/* Timeline Dot */}
                  <div
                    className={`absolute left-0 w-12 h-12 rounded-full flex items-center justify-center ${
                      milestone.status === 'completed'
                        ? 'bg-green-100'
                        : milestone.status === 'in-progress'
                        ? 'bg-blue-100'
                        : 'bg-gray-100'
                    }`}
                  >
                    {milestone.status === 'completed' && (
                      <CheckCircle size={24} className="text-green-600" />
                    )}
                    {milestone.status === 'in-progress' && (
                      <Clock size={24} className="text-blue-600" />
                    )}
                    {milestone.status === 'pending' && (
                      <Target size={24} className="text-gray-600" />
                    )}
                  </div>

                  {/* Content */}
                  <div className={`p-4 rounded-lg border ${
                    milestone.status === 'completed'
                      ? 'border-green-200 bg-green-50'
                      : milestone.status === 'in-progress'
                      ? 'border-blue-200 bg-blue-50'
                      : 'border-gray-200 bg-gray-50'
                  }`}>
                    <div className="flex items-start justify-between mb-2">
                      <div>
                        <h3 className="font-semibold text-gray-900">{milestone.title}</h3>
                        <p className="text-sm text-gray-600 mt-1">{milestone.description}</p>
                      </div>
                      <span className={`text-xs px-2 py-1 rounded whitespace-nowrap ml-2 ${
                        milestone.category === 'Educación'
                          ? 'bg-blue-200 text-blue-700'
                          : milestone.category === 'Empleabilidad'
                          ? 'bg-green-200 text-green-700'
                          : 'bg-yellow-200 text-yellow-700'
                      }`}>
                        {milestone.category}
                      </span>
                    </div>
                    <p className="text-xs text-gray-500">{milestone.date}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </CardBody>
      </Card>

      {/* Skills and Achievements */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <h2 className="text-lg font-semibold">Habilidades Desarrolladas</h2>
          </CardHeader>
          <CardBody>
            <div className="space-y-3">
              {['Panadería Artesanal', 'Trabajo en Equipo', 'Cocina Saludable', 'Gestión del Tiempo', 'Atención al Cliente'].map(
                (skill, index) => (
                  <div key={index} className="flex items-center gap-3">
                    <div className="w-2 h-2 rounded-full bg-blue-600"></div>
                    <span className="text-gray-700">{skill}</span>
                    <div className="ml-auto flex text-yellow-400">
                      {'★'.repeat(3 + (index % 2))}
                      {'☆'.repeat(2 - (index % 2))}
                    </div>
                  </div>
                )
              )}
            </div>
          </CardBody>
        </Card>

        <Card>
          <CardHeader>
            <h2 className="text-lg font-semibold">Próximas Metas</h2>
          </CardHeader>
          <CardBody>
            <div className="space-y-4">
              <div className="p-3 border-l-4 border-purple-600 bg-purple-50">
                <p className="font-medium text-gray-900">Meta 1: Completar Costura y Diseño</p>
                <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
                  <div className="bg-purple-600 h-2 rounded-full" style={{ width: '65%' }}></div>
                </div>
                <p className="text-xs text-gray-600 mt-1">65% completado</p>
              </div>

              <div className="p-3 border-l-4 border-pink-600 bg-pink-50">
                <p className="font-medium text-gray-900">Meta 2: Obtener Certificación Profesional</p>
                <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
                  <div className="bg-pink-600 h-2 rounded-full" style={{ width: '30%' }}></div>
                </div>
                <p className="text-xs text-gray-600 mt-1">30% completado</p>
              </div>

              <div className="p-3 border-l-4 border-green-600 bg-green-50">
                <p className="font-medium text-gray-900">Meta 3: Inserción Laboral</p>
                <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
                  <div className="bg-green-600 h-2 rounded-full" style={{ width: '45%' }}></div>
                </div>
                <p className="text-xs text-gray-600 mt-1">45% completado</p>
              </div>
            </div>
          </CardBody>
        </Card>
      </div>
    </div>
  );
}
