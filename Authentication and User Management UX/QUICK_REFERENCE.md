# Quick Reference - Sistema Buen Vivir Prototipo

**Consulta rápida para decisiones durante desarrollo**

---

## 🏗️ ESTRUCTURA DE CARPETAS

```
src/app/
├── components/
│   ├── base/                    ← Componentes reutilizables
│   │   ├── Modal.tsx
│   │   ├── Form.tsx
│   │   ├── DataTable.tsx
│   │   └── ...
│   ├── features/                ← Componentes por módulo
│   │   ├── auth/
│   │   ├── education/
│   │   ├── humane-promo/
│   │   ├── job-insertion/
│   │   └── campaigns/
│   ├── ui/                      ← shadcn/ui (existente)
│   └── [Page components]        ← Rutas principales
├── hooks/                       ← Lógica reutilizable
├── types/                       ← Tipos centralizados
├── constants/                   ← Valores globales
├── utils/                       ← Funciones helper
├── services/                    ← Mock APIs / servicios
├── store/                       ← State global (Zustand)
└── styles/                      ← CSS global
```

---

## 📋 NOMBRES DE COSAS

### Variables y Funciones
```typescript
userName              ✅ camelCase
get_user_name       ❌ snake_case
getUserName()       ✅ camelCase function
GET_USER_NAME       ❌ CONSTANT (solo para const)
```

### Componentes y Tipos
```typescript
UserProfile.tsx     ✅ PascalCase componente
interface User      ✅ PascalCase interface
type Role = '...'   ✅ PascalCase type
user_profile.tsx    ❌ snake_case componente
```

### Constantes
```typescript
MAX_FILE_SIZE = 5242880     ✅ UPPER_SNAKE_CASE
API_TIMEOUT = 30000         ✅ UPPER_SNAKE_CASE
maxFileSize = 5242880       ❌ camelCase
```

---

## 🔤 DATOS: ESPAÑOL vs INGLÉS

### Código TypeScript
```typescript
// Interface nombres: INGLÉS
interface Course {
  id: string;
  title: string;
  description: string;
  status: CourseStatus;
}

// Variables y funciones: INGLÉS
const getCourseById = (id: string) => {...};
const activeUsers = users.filter(u => u.status === 'active');
```

### Datos en BD / Responses
```typescript
// En respuesta de API: ESPAÑOL (se define en types/)
const course = {
  id: '1',
  titulo: 'Panadería Artesanal',
  descripcion: '...',
  estado: 'activo',
};

// Pero se mapea a types al recibir:
const typedCourse: Course = {
  id: course.id,
  title: course.titulo,
  description: course.descripcion,
  status: course.estado as CourseStatus,
};
```

---

## 🎨 COLORES Y CLASES

### Orden de clases Tailwind
```typescript
// ✅ ORDEN: layout → spacing → size → color → effects → animation
className="
  flex flex-col              // layout
  gap-4 p-6 mb-8             // spacing
  w-full max-w-4xl h-screen  // sizing
  bg-background text-foreground border border-border  // colors
  rounded-lg shadow-md hover:shadow-lg  // effects
  transition-shadow duration-300       // animation
"

// ❌ DESORDENADO
className="shadow-md w-full bg-background p-6 gap-4 text-foreground"
```

### Escala de responsive
```typescript
// Mobile first
className="
  text-base              // base: mobile
  md:text-lg             // >= 768px
  lg:text-xl             // >= 1024px
  xl:text-2xl            // >= 1280px
"
```

---

## 🛠️ COMPONENTES: CUÁNDO USAR QUÉ

### Modal / Dialog
```typescript
// Para:
- Confirmación de acción
- Formulario pequeño
- Alerta importante

// Uso:
<AlertDialog>
  <AlertDialogTrigger>Eliminar</AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogTitle>¿Estás seguro?</AlertDialogTitle>
    <AlertDialogAction onClick={onDelete}>Sí</AlertDialogAction>
  </AlertDialogContent>
</AlertDialog>
```

### Form
```typescript
// Para:
- Inputs con validación
- Cambios de estado
- Submits

// Uso:
<Form {...form}>
  <form onSubmit={form.handleSubmit(onSubmit)}>
    <FormField control={form.control} name="email" render={...} />
  </form>
</Form>
```

### DataTable
```typescript
// Para:
- Listados con muchos items
- Paginación necesaria
- Ordenamiento o filtrado

// NO para:
- Listas pequeñas < 10 items
- Lectura simple (usar ul/li)
```

### Card
```typescript
// Para:
- Agrupar contenido relacionado
- Crear secciones visuales

// Uso:
<Card>
  <CardHeader>
    <CardTitle>Título</CardTitle>
  </CardHeader>
  <CardContent>Contenido</CardContent>
</Card>
```

---

## 📝 PATRÓN DE COMPONENTE

```typescript
import { useState } from 'react';
import type { User } from '../types';
import { Button } from './ui/button';

interface MyComponentProps {
  user: User;
  onUpdate?: (user: User) => void;
}

/**
 * MyComponent - Descripción breve
 * @param {MyComponentProps} props
 * @example
 * <MyComponent user={user} onUpdate={handleUpdate} />
 */
export function MyComponent({ user, onUpdate }: MyComponentProps) {
  // 1. State
  const [isEditing, setIsEditing] = useState(false);
  
  // 2. Handlers
  const handleSave = () => {
    onUpdate?.(user);
    setIsEditing(false);
  };
  
  // 3. Render
  return (
    <div className="p-4 border rounded-lg">
      {user.name}
      <Button onClick={handleSave}>Guardar</Button>
    </div>
  );
}
```

---

## 🔗 TIPOS MÁS USADOS

```typescript
// De src/types/index.ts

// Usuario
import type { User, Role, UserStatus } from '../types';

// Cursos
import type { Course, CourseStatus, Enrollment } from '../types';

// Ofertas
import type { JobOffer, JobApplication } from '../types';

// Campañas
import type { Campaign } from '../types';

// Enums
import { Role, CourseStatus, JobOfferStatus } from '../types';
```

---

## ⚡ VALIDACIÓN RÁPIDA

### Con Zod
```typescript
import { z } from 'zod';

// Definir schema
const courseSchema = z.object({
  titulo: z.string().min(3),
  duracion: z.number().min(1),
});

// Usar en form
const result = courseSchema.safeParse(formData);
if (!result.success) {
  console.error(result.error.flatten().fieldErrors);
}
```

### Validación inline
```typescript
const validateEmail = (email: string) => 
  /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);

const validatePhone = (phone: string) => 
  /^\d{10}$/.test(phone.replace(/\D/g, ''));
```

---

## 🔄 ASYNC / AWAIT PATTERN

```typescript
const handleSubmit = async () => {
  setIsLoading(true);
  setError(null);
  
  try {
    const data = await api.post('/endpoint', payload);
    setSuccess('¡Hecho!');
  } catch (err) {
    setError(err instanceof Error ? err.message : 'Error desconocido');
  } finally {
    setIsLoading(false);
  }
};
```

---

## 📦 MOCK DATA PATTERN

```typescript
// src/services/courseService.ts

export const mockCourses: Course[] = [
  {
    id: 'course-1',
    title: 'Panadería Artesanal',
    description: '...',
    status: 'activo',
    capacity: 30,
    enrolled: 24,
    // ... resto de campos
  },
  // ... más cursos
];

export const courseService = {
  getAll: async () => mockCourses,
  getById: async (id: string) => 
    mockCourses.find(c => c.id === id) || null,
  create: async (data: CourseInput) => ({
    ...data,
    id: nanoid(),
    status: 'draft' as const,
    createdAt: new Date(),
  }),
  update: async (id: string, data: Partial<Course>) => ({
    ...mockCourses.find(c => c.id === id),
    ...data,
    updatedAt: new Date(),
  }),
};
```

---

## 🪝 CUSTOM HOOKS PATTERN

```typescript
// src/hooks/useCourses.ts

import { useState, useEffect } from 'react';
import type { Course } from '../types';
import { courseService } from '../services/courseService';

export function useCourses() {
  const [courses, setCourses] = useState<Course[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  
  useEffect(() => {
    const loadCourses = async () => {
      setIsLoading(true);
      try {
        const data = await courseService.getAll();
        setCourses(data);
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Unknown error'));
      } finally {
        setIsLoading(false);
      }
    };
    
    loadCourses();
  }, []);
  
  return { courses, isLoading, error };
}

// Uso en componente:
export function CourseList() {
  const { courses, isLoading } = useCourses();
  
  if (isLoading) return <div>Loading...</div>;
  
  return (
    <ul>
      {courses.map(course => (
        <li key={course.id}>{course.title}</li>
      ))}
    </ul>
  );
}
```

---

## 🚫 ERRORES COMUNES

| Problema | ❌ MAL | ✅ BIEN |
|----------|--------|---------|
| Type `any` | `const user: any = {...}` | `const user: User = {...}` |
| Props sin interfaz | `function Card(props)` | `function Card(props: CardProps)` |
| setState en render | En render directo | En useEffect |
| Clases desordenadas | `p-6 bg-red-500 flex` | `flex p-6 bg-red-500` |
| Sin validación | Input directo | Validar con Zod/onChange |
| Sin try-catch | `await api.call()` | `try { await } catch` |
| console.log en producción | Dejar logs | Usar logger service |
| Sin tipado | `const data = response` | `const data: Course = response` |

---

## 🧪 TEST PATTERN

```typescript
import { render, screen, fireEvent } from '@testing-library/react';
import { CourseCard } from './CourseCard';

describe('CourseCard', () => {
  it('should display course title', () => {
    const course = { id: '1', title: 'Test Course' };
    render(<CourseCard course={course} />);
    
    expect(screen.getByText('Test Course')).toBeInTheDocument();
  });
  
  it('should call onClick when button is clicked', () => {
    const onClick = jest.fn();
    render(<CourseCard course={{...}} onClick={onClick} />);
    
    fireEvent.click(screen.getByRole('button'));
    expect(onClick).toHaveBeenCalled();
  });
});
```

---

## 📊 FLUJOS CLAVE

### Autenticación
```
Login → Validar credentials → JWT + refresh token → Guardar en localStorage → Redirect a dashboard
```

### Inscripción a Curso
```
CourseCatalog → CourseDetail → Click "Inscribirse" → Validar → Crear Enrollment
→ Toast success → Notificación email (mock) → Redirect a ProgressTracking
```

### Aplicar a Oferta
```
PublicJobOffers → JobOfferDetail → Click "Aplicar" → Generar CV automático
→ Crear JobApplication → Toast + Email notification → Ver estado
```

### Crear Campaña
```
CampaignManagement → Click "Nueva campaña" → CampaignForm
→ Editor visual → Subir imágenes → Publicar → Estado "publicada" → Ver en pública
```

---

## 🎯 MÉTRICAS MÍNIMAS

```
✅ TypeScript: 0% any types en código nuevo
✅ Testing: 70%+ cobertura lógica crítica
✅ Performance: <500ms p95 en endpoints
✅ Accesibilidad: WCAG 2.1 AA
✅ Code: Seguir CONVENCIONES_CODIGO.md
✅ Validación: 100% inputs con validación
✅ Errores: 100% async con try-catch
✅ Tipos: 100% datos tipados desde src/types/
```

---

## 🔗 LINKS RÁPIDOS

- **Plan completo**: `PLAN_EXPANSION_PROTOTIPO.md`
- **Inventario**: `COMPONENTES_INVENTORY.md`
- **Tipos**: `DATOS_SCHEMA.md` → crear en `src/types/index.ts`
- **Convenciones**: `CONVENCIONES_CODIGO.md`

---

**Imprimir y pegar en monitor o tablet para consulta rápida** 🖨️
