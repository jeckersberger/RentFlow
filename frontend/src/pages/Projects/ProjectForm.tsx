import { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { ArrowLeft, Save } from 'lucide-react';
import toast from 'react-hot-toast';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import * as projectApi from '@/services/projects';
import type { Project } from '@/types/project';
import './ProjectForm.scss';

const STATUS_OPTIONS = [
  { value: 'draft', label: 'Entwurf' },
  { value: 'confirmed', label: 'Bestaetigt' },
  { value: 'active', label: 'Aktiv' },
  { value: 'completed', label: 'Abgeschlossen' },
  { value: 'cancelled', label: 'Storniert' },
  { value: 'archived', label: 'Archiviert' },
];

interface ProjectFormData {
  name: string;
  project_number: string;
  status: string;
  contact_name: string;
  contact_email: string;
  contact_phone: string;
  venue_name: string;
  venue_address: string;
  start_date: string;
  end_date: string;
  budget: string;
  notes: string;
}

function centsToEur(cents: number | undefined): string {
  if (cents == null || cents === 0) return '';
  return (cents / 100).toFixed(2);
}

function eurToCents(eurStr: string): number {
  const parsed = parseFloat(eurStr.replace(',', '.'));
  if (isNaN(parsed)) return 0;
  return Math.round(parsed * 100);
}

function toDateInput(dateStr?: string): string {
  if (!dateStr) return '';
  return dateStr.slice(0, 10);
}

export default function ProjectForm() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isEdit = !!id;

  const { data: existing, isLoading } = useQuery({
    queryKey: ['projects', id],
    queryFn: () => projectApi.get(id!),
    enabled: isEdit,
  });

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isDirty },
  } = useForm<ProjectFormData>({
    defaultValues: {
      name: '',
      project_number: '',
      status: 'draft',
      contact_name: '',
      contact_email: '',
      contact_phone: '',
      venue_name: '',
      venue_address: '',
      start_date: '',
      end_date: '',
      budget: '',
      notes: '',
    },
  });

  useEffect(() => {
    if (existing) {
      reset({
        name: existing.name || '',
        project_number: existing.project_number || '',
        status: existing.status || 'draft',
        contact_name: existing.contact_name || '',
        contact_email: existing.contact_email || '',
        contact_phone: existing.contact_phone || '',
        venue_name: existing.venue_name || '',
        venue_address: existing.venue_address || '',
        start_date: toDateInput(existing.start_date),
        end_date: toDateInput(existing.end_date),
        budget: centsToEur(existing.budget),
        notes: existing.notes || '',
      });
    }
  }, [existing, reset]);

  const mutation = useMutation({
    mutationFn: (payload: Partial<Project>) =>
      isEdit ? projectApi.update(id!, payload) : projectApi.create(payload),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      toast.success(isEdit ? 'Projekt aktualisiert' : 'Projekt erstellt');
      navigate(`/projects/${result.id}`);
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Fehler beim Speichern');
    },
  });

  const onSubmit = (formData: ProjectFormData) => {
    const payload: Partial<Project> = {
      name: formData.name,
      project_number: formData.project_number,
      status: formData.status,
      contact_name: formData.contact_name,
      contact_email: formData.contact_email,
      contact_phone: formData.contact_phone,
      venue_name: formData.venue_name,
      venue_address: formData.venue_address,
      start_date: formData.start_date || undefined,
      end_date: formData.end_date || undefined,
      budget: eurToCents(formData.budget),
      notes: formData.notes,
    };
    mutation.mutate(payload);
  };

  if (isEdit && isLoading) {
    return (
      <PageWrapper title="Projekt">
        <div className="detail-skeleton">
          <div className="skeleton skeleton--heading" />
          <div className="skeleton skeleton--block" />
        </div>
      </PageWrapper>
    );
  }

  return (
    <PageWrapper title={isEdit ? 'Projekt bearbeiten' : 'Neues Projekt'}>
      <button
        type="button"
        className="btn btn--ghost"
        onClick={() => navigate(-1)}
      >
        <ArrowLeft size={18} />
        <span>Zurueck</span>
      </button>

      <motion.div
        className="form-card"
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <form onSubmit={handleSubmit(onSubmit)} noValidate>
          <div className="form-grid">
            <div className="form-group">
              <label htmlFor="pf-name">Name *</label>
              <input
                id="pf-name"
                type="text"
                placeholder="Projektname"
                {...register('name', { required: 'Name ist erforderlich' })}
              />
              {errors.name && (
                <span className="form-error">{errors.name.message}</span>
              )}
            </div>

            <div className="form-group">
              <label htmlFor="pf-number">Projektnummer</label>
              <input
                id="pf-number"
                type="text"
                placeholder="z.B. P-2026-001"
                {...register('project_number')}
              />
            </div>

            <div className="form-group">
              <label htmlFor="pf-status">Status</label>
              <select id="pf-status" {...register('status')}>
                {STATUS_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>

            <div className="form-group">
              <label htmlFor="pf-budget">Budget (EUR)</label>
              <input
                id="pf-budget"
                type="number"
                step="0.01"
                min="0"
                placeholder="0.00"
                {...register('budget')}
              />
            </div>

            <div className="form-group">
              <label htmlFor="pf-contact-name">Ansprechpartner</label>
              <input
                id="pf-contact-name"
                type="text"
                placeholder="Name"
                {...register('contact_name')}
              />
            </div>

            <div className="form-group">
              <label htmlFor="pf-contact-email">E-Mail</label>
              <input
                id="pf-contact-email"
                type="email"
                placeholder="name@firma.de"
                {...register('contact_email', {
                  pattern: {
                    value: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
                    message: 'Ungueltige E-Mail-Adresse',
                  },
                })}
              />
              {errors.contact_email && (
                <span className="form-error">
                  {errors.contact_email.message}
                </span>
              )}
            </div>

            <div className="form-group">
              <label htmlFor="pf-contact-phone">Telefon</label>
              <input
                id="pf-contact-phone"
                type="text"
                placeholder="+49 ..."
                {...register('contact_phone')}
              />
            </div>

            <div className="form-group">
              <label htmlFor="pf-venue">Veranstaltungsort</label>
              <input
                id="pf-venue"
                type="text"
                placeholder="Location"
                {...register('venue_name')}
              />
            </div>

            <div className="form-group form-group--full">
              <label htmlFor="pf-venue-address">Adresse</label>
              <input
                id="pf-venue-address"
                type="text"
                placeholder="Strasse, PLZ Ort"
                {...register('venue_address')}
              />
            </div>

            <div className="form-group">
              <label htmlFor="pf-start">Startdatum</label>
              <input
                id="pf-start"
                type="date"
                {...register('start_date')}
              />
            </div>

            <div className="form-group">
              <label htmlFor="pf-end">Enddatum</label>
              <input
                id="pf-end"
                type="date"
                {...register('end_date')}
              />
            </div>

            <div className="form-group form-group--full">
              <label htmlFor="pf-notes">Notizen</label>
              <textarea
                id="pf-notes"
                placeholder="Zusaetzliche Informationen..."
                rows={3}
                {...register('notes')}
              />
            </div>
          </div>

          <div className="form-actions">
            <button
              type="button"
              className="btn btn--secondary"
              onClick={() => navigate(-1)}
            >
              Abbrechen
            </button>
            <button
              type="submit"
              className="btn btn--primary"
              disabled={mutation.isPending || (!isDirty && isEdit)}
            >
              {mutation.isPending ? (
                <span className="loading-spinner loading-spinner--small" />
              ) : (
                <>
                  <Save size={16} />
                  <span>{isEdit ? 'Speichern' : 'Erstellen'}</span>
                </>
              )}
            </button>
          </div>
        </form>
      </motion.div>
    </PageWrapper>
  );
}
