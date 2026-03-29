import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { ChevronLeft, ChevronRight, Download } from 'lucide-react';
import { motion } from 'framer-motion';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import * as projectApi from '@/services/projects';
import type { CalendarEvent } from '@/services/projects';
import './CalendarPage.scss';

const WEEKDAYS = ['Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa', 'So'];

function pad(n: number): string {
  return n.toString().padStart(2, '0');
}

function formatMonth(year: number, month: number): string {
  const date = new Date(year, month);
  return date.toLocaleDateString('de-DE', { month: 'long', year: 'numeric' });
}

function dateStr(year: number, month: number, day: number): string {
  return `${year}-${pad(month + 1)}-${pad(day)}`;
}

function getDaysInMonth(year: number, month: number): number {
  return new Date(year, month + 1, 0).getDate();
}

/** Returns 0=Mon ... 6=Sun for the first day of the month. */
function getStartDayOfWeek(year: number, month: number): number {
  const jsDay = new Date(year, month, 1).getDay(); // 0=Sun
  return jsDay === 0 ? 6 : jsDay - 1;
}

interface CalendarDay {
  day: number;
  dateKey: string;
  isCurrentMonth: boolean;
  isToday: boolean;
}

function buildCalendarGrid(year: number, month: number): CalendarDay[] {
  const today = new Date();
  const todayKey = dateStr(today.getFullYear(), today.getMonth(), today.getDate());

  const daysInMonth = getDaysInMonth(year, month);
  const startDay = getStartDayOfWeek(year, month);

  // Previous month fill
  const prevMonth = month === 0 ? 11 : month - 1;
  const prevYear = month === 0 ? year - 1 : year;
  const daysInPrevMonth = getDaysInMonth(prevYear, prevMonth);

  const grid: CalendarDay[] = [];

  for (let i = startDay - 1; i >= 0; i--) {
    const day = daysInPrevMonth - i;
    grid.push({
      day,
      dateKey: dateStr(prevYear, prevMonth, day),
      isCurrentMonth: false,
      isToday: false,
    });
  }

  for (let d = 1; d <= daysInMonth; d++) {
    const key = dateStr(year, month, d);
    grid.push({
      day: d,
      dateKey: key,
      isCurrentMonth: true,
      isToday: key === todayKey,
    });
  }

  // Next month fill to complete the grid (always 6 rows = 42 cells)
  const nextMonth = month === 11 ? 0 : month + 1;
  const nextYear = month === 11 ? year + 1 : year;
  const remaining = 42 - grid.length;
  for (let d = 1; d <= remaining; d++) {
    grid.push({
      day: d,
      dateKey: dateStr(nextYear, nextMonth, d),
      isCurrentMonth: false,
      isToday: false,
    });
  }

  return grid;
}

function getEventsForDay(
  events: CalendarEvent[],
  dateKey: string
): CalendarEvent[] {
  return events.filter((ev) => {
    if (!ev.start_date) return false;
    const start = ev.start_date;
    const end = ev.end_date || ev.start_date;
    return dateKey >= start && dateKey <= end;
  });
}

export default function CalendarPage() {
  const navigate = useNavigate();
  const now = new Date();
  const [year, setYear] = useState(now.getFullYear());
  const [month, setMonth] = useState(now.getMonth());

  const from = dateStr(year, month, 1);
  const toDate = new Date(year, month + 1, 0);
  const to = dateStr(toDate.getFullYear(), toDate.getMonth(), toDate.getDate());

  const { data: events = [], isLoading } = useQuery({
    queryKey: ['calendar', from, to],
    queryFn: () => projectApi.getCalendar(from, to),
  });

  const calendarEvents = Array.isArray(events) ? events : [];

  const grid = useMemo(() => buildCalendarGrid(year, month), [year, month]);

  const goPrev = () => {
    if (month === 0) {
      setYear(year - 1);
      setMonth(11);
    } else {
      setMonth(month - 1);
    }
  };

  const goNext = () => {
    if (month === 11) {
      setYear(year + 1);
      setMonth(0);
    } else {
      setMonth(month + 1);
    }
  };

  const goToday = () => {
    setYear(now.getFullYear());
    setMonth(now.getMonth());
  };

  const handleExportICS = () => {
    const token = localStorage.getItem('cd_access_token') || '';
    const url = `/api/v1/projects/calendar.ics?from=${from}&to=${to}`;
    // Use a temporary anchor to trigger the download with auth
    fetch(url, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((res) => res.blob())
      .then((blob) => {
        const a = document.createElement('a');
        a.href = URL.createObjectURL(blob);
        a.download = 'cratedesk-calendar.ics';
        a.click();
        URL.revokeObjectURL(a.href);
      });
  };

  const headerActions = (
    <div className="calendar-actions">
      <button className="btn btn--ghost" onClick={goPrev} title="Vorheriger Monat">
        <ChevronLeft size={18} />
      </button>
      <button className="btn btn--ghost calendar-actions__today" onClick={goToday}>
        Heute
      </button>
      <button className="btn btn--ghost" onClick={goNext} title="Naechster Monat">
        <ChevronRight size={18} />
      </button>
      <span className="calendar-actions__month">{formatMonth(year, month)}</span>
      <button className="btn btn--secondary" onClick={handleExportICS} title="ICS Export">
        <Download size={16} />
        <span>ICS</span>
      </button>
    </div>
  );

  return (
    <PageWrapper title="Kalender" actions={headerActions}>
      {isLoading ? (
        <div className="calendar-loading">Lade Kalender...</div>
      ) : (
        <motion.div
          className="calendar-grid-wrapper"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.2 }}
        >
          <div className="calendar-grid">
            {WEEKDAYS.map((wd) => (
              <div key={wd} className="calendar-grid__header">
                {wd}
              </div>
            ))}
            {grid.map((cell) => {
              const dayEvents = getEventsForDay(calendarEvents, cell.dateKey);
              return (
                <div
                  key={cell.dateKey}
                  className={[
                    'calendar-grid__cell',
                    !cell.isCurrentMonth && 'calendar-grid__cell--outside',
                    cell.isToday && 'calendar-grid__cell--today',
                  ]
                    .filter(Boolean)
                    .join(' ')}
                >
                  <span className="calendar-grid__day">{cell.day}</span>
                  <div className="calendar-grid__events">
                    {dayEvents.slice(0, 3).map((ev) => (
                      <button
                        key={ev.id}
                        className="calendar-event"
                        style={{ backgroundColor: ev.color || '#3b82f6' }}
                        onClick={() => navigate(`/projects/${ev.id}`)}
                        title={`${ev.title}${ev.venue_name ? ' – ' + ev.venue_name : ''}`}
                      >
                        <span className="calendar-event__title">{ev.title}</span>
                      </button>
                    ))}
                    {dayEvents.length > 3 && (
                      <span className="calendar-grid__more">
                        +{dayEvents.length - 3}
                      </span>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </motion.div>
      )}
    </PageWrapper>
  );
}
