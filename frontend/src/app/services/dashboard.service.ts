import { Injectable, inject } from "@angular/core";
import { Observable, catchError, forkJoin, map, of } from "rxjs";
import {
  CursoOrigen,
  DashboardData,
  FiltrosDashboard,
  Grafica,
  Indicador,
  IndicadorId,
  ModuloDashboard,
  ParticipanteHabilidadOrigen,
  ParticipanteOrigen,
  PostulacionOrigen,
  PromocionOrigen,
  PuntoSerie,
} from "../core/models/dashboard.model";
import { DashboardSourcesService } from "./dashboard-sources.service";
import {
  contarPorMes,
  enRango,
  estadoCurso,
  fichaActivaEn,
  finDelDia,
  habilidadAceptada,
  insercionLaboral,
  mesesDelPeriodo,
  participanteActivoEn,
  participantesConSeguimiento,
} from "./dashboard.calculos";

/** Datos crudos de cada fuente; null = el servicio no respondió. */
export interface DatosOrigen {
  participantes: ParticipanteOrigen[] | null;
  cursos: CursoOrigen[] | null;
  habilidades: ParticipanteHabilidadOrigen[] | null;
  postulaciones: PostulacionOrigen[] | null;
  promocion: PromocionOrigen | null;
}

type MetaIndicador = Omit<Indicador, "valor" | "disponible">;

/**
 * Indicadores, en el orden en que se ven: los cinco de HU-52 y los dos de
 * Promoción Humana que pidió el equipo (el DER prevé que el Dashboard
 * consulte también a ese servicio).
 */
export const INDICADORES: MetaIndicador[] = [
  {
    id: "participantesActivos",
    titulo: "Participantes activos",
    descripcion: "Participantes con la cuenta activa al cierre del período.",
    modulo: "usuarios",
    fuente: "autenticacion",
  },
  {
    id: "cursosEnDesarrollo",
    titulo: "Cursos en desarrollo",
    descripcion: "Cursos publicados con actividades pendientes.",
    modulo: "educacion",
    fuente: "educacion",
  },
  {
    id: "cursosFinalizados",
    titulo: "Cursos finalizados",
    descripcion: "Cursos publicados cuyas actividades ya ocurrieron todas.",
    modulo: "educacion",
    fuente: "educacion",
  },
  {
    id: "habilidadesDesarrolladas",
    titulo: "Habilidades desarrolladas",
    descripcion: "Habilidades aceptadas por la administración en el período.",
    modulo: "educacion",
    fuente: "educacion",
  },
  {
    id: "insercionesLaborales",
    titulo: "Inserciones laborales",
    descripcion: "Postulaciones que pasaron a «seleccionado» en el período.",
    modulo: "insercion",
    fuente: "insercion",
  },
  {
    id: "fichasActivas",
    titulo: "Fichas activas",
    descripcion: "Fichas de proceso personal abiertas al cierre del período.",
    modulo: "promocion",
    fuente: "promocion",
  },
  {
    id: "participantesConSeguimiento",
    titulo: "Participantes con seguimiento",
    descripcion: "Participantes con al menos un hito de avance en el período.",
    modulo: "promocion",
    fuente: "promocion",
  },
];

/**
 * Indicadores y gráficas del dashboard (HU-52, HU-53, HU-54).
 *
 * Pide los datos a cada servicio de origen por separado: si uno falla,
 * sus indicadores salen como "No disponible" y el resto se muestra igual.
 */
@Injectable({ providedIn: "root" })
export class DashboardService {
  private sources = inject(DashboardSourcesService);

  areas() {
    return this.sources.areas();
  }

  cargar(filtros: FiltrosDashboard): Observable<DashboardData> {
    const seguro = <T>(obs: Observable<T>) =>
      obs.pipe(catchError(() => of(null)));
    return forkJoin({
      participantes: seguro(this.sources.participantes()),
      cursos: seguro(this.sources.cursos()),
      habilidades: seguro(this.sources.habilidades()),
      postulaciones: seguro(this.sources.postulaciones()),
      promocion: seguro(this.sources.promocion()),
    }).pipe(map((datos) => calcularDashboard(datos, filtros)));
  }
}

/** Indicadores y gráficas para los filtros dados. Pura: se prueba sola. */
export function calcularDashboard(
  datos: DatosOrigen,
  filtros: FiltrosDashboard,
): DashboardData {
  const { desde, hasta, areaId } = filtros;
  const corte = finDelDia(hasta);
  const meses = mesesDelPeriodo(desde, hasta);

  // El área formativa afecta solo a cursos y habilidades.
  const cursos = datos.cursos?.filter((c) => !areaId || c.areaId === areaId);
  const habilidades = datos.habilidades?.filter(
    (h) => habilidadAceptada(h) && (!areaId || h.areaId === areaId),
  );
  const inserciones = datos.postulaciones?.filter(insercionLaboral);
  // El área formativa no aplica a Promoción Humana.
  const promocion = datos.promocion;

  const cursosEn = (ref: Date, estado: "en_desarrollo" | "finalizado") =>
    (cursos ?? []).filter((c) => estadoCurso(c, ref) === estado).length;

  const valores: Record<IndicadorId, number | null> = {
    participantesActivos: datos.participantes
      ? datos.participantes.filter((p) => participanteActivoEn(p, corte)).length
      : null,
    cursosEnDesarrollo: cursos ? cursosEn(corte, "en_desarrollo") : null,
    cursosFinalizados: cursos ? cursosEn(corte, "finalizado") : null,
    habilidadesDesarrolladas: habilidades
      ? habilidades.filter((h) => enRango(h.fecha, desde, hasta)).length
      : null,
    insercionesLaborales: inserciones
      ? inserciones.filter((p) => enRango(p.fechaEstado, desde, hasta)).length
      : null,
    fichasActivas: promocion
      ? promocion.fichas.filter((f) => fichaActivaEn(f, corte)).length
      : null,
    participantesConSeguimiento: promocion
      ? participantesConSeguimiento(promocion.hitos, desde, hasta)
      : null,
  };

  const indicadores: Indicador[] = INDICADORES.map((meta) => ({
    ...meta,
    valor: valores[meta.id],
    disponible: valores[meta.id] !== null,
  }));

  const porMes = (fn: (corteMes: Date) => number): PuntoSerie[] =>
    meses.map(({ mes, corte: c }) => ({ mes, valor: fn(c) }));

  const graficas: Grafica[] = [
    {
      id: "participantes",
      titulo: "Participantes activos",
      tipo: "line",
      modulo: "usuarios",
      indicadores: ["participantesActivos"],
      series: datos.participantes
        ? [
            {
              nombre: "Participantes activos",
              puntos: porMes(
                (c) =>
                  datos.participantes!.filter((p) => participanteActivoEn(p, c))
                    .length,
              ),
            },
          ]
        : null,
    },
    {
      id: "cursos",
      titulo: "Cursos en desarrollo y finalizados",
      tipo: "bar",
      modulo: "educacion",
      indicadores: ["cursosEnDesarrollo", "cursosFinalizados"],
      series: cursos
        ? [
            {
              nombre: "En desarrollo",
              puntos: porMes((c) => cursosEn(c, "en_desarrollo")),
            },
            {
              nombre: "Finalizados",
              puntos: porMes((c) => cursosEn(c, "finalizado")),
            },
          ]
        : null,
    },
    {
      id: "habilidades",
      titulo: "Habilidades desarrolladas por mes",
      tipo: "line",
      modulo: "educacion",
      indicadores: ["habilidadesDesarrolladas"],
      series: habilidades
        ? [
            {
              nombre: "Habilidades desarrolladas",
              puntos: contarPorMes(habilidades, (h) => h.fecha, desde, hasta),
            },
          ]
        : null,
    },
    {
      id: "inserciones",
      titulo: "Inserciones laborales por mes",
      tipo: "line",
      modulo: "insercion",
      indicadores: ["insercionesLaborales"],
      series: inserciones
        ? [
            {
              nombre: "Inserciones laborales",
              puntos: contarPorMes(
                inserciones,
                (p) => p.fechaEstado,
                desde,
                hasta,
              ),
            },
          ]
        : null,
    },
    {
      id: "promocion",
      titulo: "Fichas activas y participantes con seguimiento",
      tipo: "bar",
      modulo: "promocion",
      indicadores: ["fichasActivas", "participantesConSeguimiento"],
      series: promocion
        ? [
            {
              nombre: "Fichas activas",
              puntos: porMes(
                (c) => promocion.fichas.filter((f) => fichaActivaEn(f, c)).length,
              ),
            },
            {
              nombre: "Con seguimiento",
              // Cada mes: participantes con algún hito en ese mes, sin
              // salirse del período elegido.
              puntos: meses.map(({ mes, corte: c }) => {
                const inicioMes = new Date(c.getFullYear(), c.getMonth(), 1);
                const ini = inicioMes < desde ? desde : inicioMes;
                return { mes, valor: participantesConSeguimiento(promocion.hitos, ini, c) };
              }),
            },
          ]
        : null,
    },
  ];

  return {
    indicadores: indicadores.filter((i) =>
      visibleEnModulo(i.modulo, filtros.modulo),
    ),
    graficas: graficas.filter((g) => visibleEnModulo(g.modulo, filtros.modulo)),
  };
}

export function visibleEnModulo(
  modulo: Exclude<ModuloDashboard, "todos">,
  filtro: ModuloDashboard,
): boolean {
  return filtro === "todos" || filtro === modulo;
}
