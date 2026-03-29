import { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, Save } from 'lucide-react';
import { motion } from 'framer-motion';
import toast from 'react-hot-toast';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import * as equipmentApi from '@/services/equipment';
import './EquipmentForm.scss';

interface EquipmentFormValues {
  name: string;
  category_id: string;
  barcode: string;
  serial_number: string;
  status: string;
  condition: string;
  rental_price_day_eur: string;
  manufacturer: string;
  model: string;
  notes: string;
}

const STATUS_OPTIONS = [
  { value: 'available', label: 'Verfuegbar' },
  { value: 'reserved', label: 'Reserviert' },
  { value: 'checked_out', label: 'Ausgegeben' },
  { value: 'in_maintenance', label: 'In Wartung' },
  { value: 'damaged', label: 'Beschaedigt' },
  { value: 'retired', label: 'Ausgemustert' },
];

const CONDITION_OPTIONS = [
  { value: 'operational', label: 'Einsatzbereit' },
  { value: 'good', label: 'Gut' },
  { value: 'fair', label: 'Akzeptabel' },
  { value: 'damaged', label: 'Beschaedigt' },
  { value: 'decommissioned', label: 'Ausser Betrieb' },
];

function centsToEur(cents: number | undefined): string {
  if (cents == null || cents === 0) return '';
  return (cents / 100).toFixed(2);
}

function eurToCents(eur: string): number {
  const parsed = parseFloat(eur.replace(',', '.'));
  if (isNaN(parsed)) return 0;
  return Math.round(parsed * 100);
}

export default function EquipmentForm() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isEdit = !!id;

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<EquipmentFormValues>({
    defaultValues: {
      name: '',
      category_id: '',
      barcode: '',
      serial_number: '',
      status: 'available',
      condition: 'operational',
      rental_price_day_eur: '',
      manufacturer: '',
      model: '',
      notes: '',
    },
  });

  const { data: existing, isLoading } = useQuery({
    queryKey: ['equipment', id],
    queryFn: () => equipmentApi.get(id!),
    enabled: isEdit,
  });

  useEffect(() => {
    if (existing) {
      reset({
        name: existing.name || '',
        category_id: existing.category_id || '',
        barcode: existing.barcode || '',
        serial_number: existing.serial_number || '',
        status: existing.status || 'available',
        condition: existing.condition || 'operational',
        rental_price_day_eur: centsToEur(existing.rental_price_day),
        manufacturer: existing.manufacturer || '',
        model: existing.model || '',
        notes: existing.notes || '',
      });
    }
  }, [existing, reset]);

  const createMutation = useMutation({
    mutationFn: (data: Partial<Record<string, unknown>>) =>
      equipmentApi.create(data as Parameters<typeof equipmentApi.create>[0]),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['equipment'] });
      toast.success('Equipment erstellt');
      navigate('/equipment');
    },
    onError: () => {
      toast.error('Fehler beim Erstellen');
    },
  });

  const updateMutation = useMutation({
    mutationFn: (data: Partial<Record<string, unknown>>) =>
      equipmentApi.update(id!, data as Parameters<typeof equipmentApi.update>[1]),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['equipment'] });
      queryClient.invalidateQueries({ queryKey: ['equipment', id] });
      toast.success('Equipment gespeichert');
      navigate('/equipment');
    },
    onError: () => {
      toast.error('Fehler beim Speichern');
    },
  });

  const onSubmit = (values: EquipmentFormValues) => {
    const payload = {
      name: values.name,
      category_id: values.category_id || undefined,
      barcode: values.barcode || undefined,
      serial_number: values.serial_number || undefined,
      status: values.status,
      condition: values.condition,
      rental_price_day: eurToCents(values.rental_price_day_eur),
      manufacturer: values.manufacturer || undefined,
      model: values.model || undefined,
      notes: values.notes || undefined,
    };

    if (isEdit) {
      updateMutation.mutate(payload);
    } else {
      createMutation.mutate(payload);
    }
  };

  const isSaving = createMutation.isPending || updateMutation.isPending;

  if (isEdit && isLoading) {
    return (
      <PageWrapper title="Equipment">
        <div className="form-skeleton">
          <div className="skeleton skeleton--heading" />
          <div className="skeleton skeleton--block" />
        </div>
      </PageWrapper>
    );
  }

  const backButton = (
    <button
      type="button"
      className="btn btn--ghost"
      onClick={() => navigate('/equipment')}
    >
      <ArrowLeft size={18} />
      <span>Zurueck</span>
    </button>
  );

  return (
    <PageWrapper title={isEdit ? 'Equipment bearbeiten' : 'Neues Equipment'}>
      {backButton}

      <motion.div
        className="form-page"
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <form onSubmit={handleSubmit(onSubmit)}>
          <div className="form-card">
            <h2 className="form-title">
              {isEdit ? 'Equipment bearbeiten' : 'Equipment anlegen'}
            </h2>

            <div className="form-grid">
              {/* Name */}
              <div className="form-group">
                <label htmlFor="name">Name *</label>
                <input
                  id="name"
                  type="text"
                  placeholder="z.B. Shure SM58"
                  {...register('name', { required: 'Name ist erforderlich' })}
                />
                {errors.name && (
                  <span className="form-error">{errors.name.message}</span>
                )}
              </div>

              {/* Category */}
              <div className="form-group">
                <label htmlFor="category_id">Kategorie</label>
                <input
                  id="category_id"
                  type="text"
                  placeholder="Kategorie-ID"
                  {...register('category_id')}
                />
              </div>

              {/* Barcode */}
              <div className="form-group">
                <label htmlFor="barcode">Barcode</label>
                <input
                  id="barcode"
                  type="text"
                  placeholder="EAN / interner Code"
                  {...register('barcode')}
                />
              </div>

              {/* Serial Number */}
              <div className="form-group">
                <label htmlFor="serial_number">Seriennummer</label>
                <input
                  id="serial_number"
                  type="text"
                  placeholder="SN-..."
                  {...register('serial_number')}
                />
              </div>

              {/* Status */}
              <div className="form-group">
                <label htmlFor="status">Status</label>
                <select id="status" {...register('status')}>
                  {STATUS_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>

              {/* Condition */}
              <div className="form-group">
                <label htmlFor="condition">Zustand</label>
                <select id="condition" {...register('condition')}>
                  {CONDITION_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>

              {/* Rental Price */}
              <div className="form-group">
                <label htmlFor="rental_price_day_eur">Tagessatz (EUR)</label>
                <input
                  id="rental_price_day_eur"
                  type="number"
                  step="0.01"
                  min="0"
                  placeholder="0.00"
                  {...register('rental_price_day_eur')}
                />
                <span className="form-hint">Eingabe in Euro, Speicherung in Cent</span>
              </div>

              {/* Manufacturer */}
              <div className="form-group">
                <label htmlFor="manufacturer">Hersteller</label>
                <input
                  id="manufacturer"
                  type="text"
                  placeholder="z.B. Shure"
                  {...register('manufacturer')}
                />
              </div>

              {/* Model */}
              <div className="form-group">
                <label htmlFor="model">Modell</label>
                <input
                  id="model"
                  type="text"
                  placeholder="z.B. SM58"
                  {...register('model')}
                />
              </div>

              {/* Notes */}
              <div className="form-group form-group--full">
                <label htmlFor="notes">Notizen</label>
                <textarea
                  id="notes"
                  rows={3}
                  placeholder="Optionale Anmerkungen..."
                  {...register('notes')}
                />
              </div>
            </div>

            <div className="form-actions">
              <button
                type="button"
                className="btn btn--secondary"
                onClick={() => navigate('/equipment')}
              >
                Abbrechen
              </button>
              <button
                type="submit"
                className="btn btn--primary"
                disabled={isSubmitting || isSaving}
              >
                <Save size={16} />
                <span>{isSaving ? 'Speichert...' : 'Speichern'}</span>
              </button>
            </div>
          </div>
        </form>
      </motion.div>
    </PageWrapper>
  );
}
