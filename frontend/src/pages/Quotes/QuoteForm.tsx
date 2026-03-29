import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useForm, useFieldArray } from 'react-hook-form';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { ArrowLeft, Save, Plus, X } from 'lucide-react';
import toast from 'react-hot-toast';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import * as quoteApi from '@/services/quotes';
import type { Quote } from '@/types/quote';
import './QuoteForm.scss';

interface QuoteItemRow {
  description: string;
  quantity: string;
  unit: string;
  unit_price: string;
}

interface QuoteFormData {
  customer_name: string;
  customer_email: string;
  customer_address: string;
  subject: string;
  quote_date: string;
  valid_until: string;
  intro_text: string;
  outro_text: string;
  kleinunternehmer: boolean;
  notes: string;
  items: QuoteItemRow[];
}

function todayStr(): string {
  return new Date().toISOString().slice(0, 10);
}

function plus30Days(): string {
  const d = new Date();
  d.setDate(d.getDate() + 30);
  return d.toISOString().slice(0, 10);
}

function toDateInput(dateStr?: string): string {
  if (!dateStr) return '';
  return dateStr.slice(0, 10);
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

function formatEurDisplay(cents: number): string {
  return new Intl.NumberFormat('de-DE', {
    style: 'currency',
    currency: 'EUR',
  }).format(cents / 100);
}

function emptyItem(): QuoteItemRow {
  return { description: '', quantity: '1', unit: 'Stueck', unit_price: '' };
}

export default function QuoteForm() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isEdit = !!id;

  const { data: existing, isLoading } = useQuery({
    queryKey: ['quotes', id],
    queryFn: () => quoteApi.get(id!),
    enabled: isEdit,
  });

  const { data: existingItems } = useQuery({
    queryKey: ['quotes', id, 'items'],
    queryFn: () => quoteApi.listItems(id!),
    enabled: isEdit,
  });

  const {
    register,
    control,
    handleSubmit,
    reset,
    watch,
    formState: { errors, isDirty },
  } = useForm<QuoteFormData>({
    defaultValues: {
      customer_name: '',
      customer_email: '',
      customer_address: '',
      subject: '',
      quote_date: todayStr(),
      valid_until: plus30Days(),
      intro_text: 'Gerne unterbreiten wir Ihnen folgendes Angebot:',
      outro_text: 'Wir freuen uns auf Ihre Rueckmeldung.',
      kleinunternehmer: true,
      notes: '',
      items: [emptyItem()],
    },
  });

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'items',
  });

  const [initialized, setInitialized] = useState(false);

  useEffect(() => {
    if (existing && !initialized) {
      const itemRows: QuoteItemRow[] =
        Array.isArray(existingItems) && existingItems.length > 0
          ? existingItems.map((it) => ({
              description: it.description || '',
              quantity: String(it.quantity || 1),
              unit: it.unit || 'Stueck',
              unit_price: centsToEur(it.unit_price),
            }))
          : [emptyItem()];

      reset({
        customer_name: existing.customer_name || '',
        customer_email: existing.customer_email || '',
        customer_address: existing.customer_address || '',
        subject: existing.subject || '',
        quote_date: toDateInput(existing.quote_date) || todayStr(),
        valid_until: toDateInput(existing.valid_until) || plus30Days(),
        intro_text: existing.intro_text || 'Gerne unterbreiten wir Ihnen folgendes Angebot:',
        outro_text: existing.outro_text || 'Wir freuen uns auf Ihre Rueckmeldung.',
        kleinunternehmer: existing.kleinunternehmer ?? true,
        notes: existing.notes || '',
        items: itemRows,
      });
      setInitialized(true);
    }
  }, [existing, existingItems, initialized, reset]);

  const watchedItems = watch('items');

  const totalCents = (watchedItems || []).reduce((sum, item) => {
    const qty = parseFloat(item.quantity) || 0;
    const price = eurToCents(item.unit_price || '0');
    return sum + qty * price;
  }, 0);

  const mutation = useMutation({
    mutationFn: (payload: Partial<Quote> & { items?: unknown[] }) =>
      isEdit ? quoteApi.update(id!, payload) : quoteApi.create(payload),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['quotes'] });
      toast.success(isEdit ? 'Angebot aktualisiert' : 'Angebot erstellt');
      navigate(`/quotes/${result.id}`);
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Fehler beim Speichern');
    },
  });

  const onSubmit = (formData: QuoteFormData) => {
    const items = formData.items
      .filter((it) => it.description.trim() !== '')
      .map((it, idx) => ({
        description: it.description,
        quantity: parseFloat(it.quantity) || 1,
        unit: it.unit || 'Stueck',
        unit_price: eurToCents(it.unit_price),
        position: idx + 1,
      }));

    const payload = {
      customer_name: formData.customer_name,
      customer_email: formData.customer_email,
      customer_address: formData.customer_address,
      subject: formData.subject,
      quote_date: formData.quote_date,
      valid_until: formData.valid_until,
      intro_text: formData.intro_text,
      outro_text: formData.outro_text,
      kleinunternehmer: formData.kleinunternehmer,
      notes: formData.notes,
      items,
    };

    mutation.mutate(payload);
  };

  if (isEdit && isLoading) {
    return (
      <PageWrapper title="Angebot">
        <div className="detail-skeleton">
          <div className="skeleton skeleton--heading" />
          <div className="skeleton skeleton--block" />
        </div>
      </PageWrapper>
    );
  }

  return (
    <PageWrapper title={isEdit ? 'Angebot bearbeiten' : 'Neues Angebot'}>
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
              <label htmlFor="qf-customer-name">Kundenname *</label>
              <input
                id="qf-customer-name"
                type="text"
                placeholder="Firmenname oder Privatperson"
                {...register('customer_name', {
                  required: 'Kundenname ist erforderlich',
                })}
              />
              {errors.customer_name && (
                <span className="form-error">
                  {errors.customer_name.message}
                </span>
              )}
            </div>

            <div className="form-group">
              <label htmlFor="qf-customer-email">E-Mail</label>
              <input
                id="qf-customer-email"
                type="email"
                placeholder="kunde@firma.de"
                {...register('customer_email', {
                  pattern: {
                    value: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
                    message: 'Ungueltige E-Mail-Adresse',
                  },
                })}
              />
              {errors.customer_email && (
                <span className="form-error">
                  {errors.customer_email.message}
                </span>
              )}
            </div>

            <div className="form-group form-group--full">
              <label htmlFor="qf-customer-address">Kundenadresse</label>
              <textarea
                id="qf-customer-address"
                placeholder="Strasse, PLZ Ort"
                rows={2}
                {...register('customer_address')}
              />
            </div>

            <div className="form-group form-group--full">
              <label htmlFor="qf-subject">Betreff</label>
              <input
                id="qf-subject"
                type="text"
                placeholder="z.B. Angebot PA-System Hochzeit Mueller"
                {...register('subject')}
              />
            </div>

            <div className="form-group">
              <label htmlFor="qf-quote-date">Angebotsdatum</label>
              <input
                id="qf-quote-date"
                type="date"
                {...register('quote_date')}
              />
            </div>

            <div className="form-group">
              <label htmlFor="qf-valid-until">Gueltig bis</label>
              <input
                id="qf-valid-until"
                type="date"
                {...register('valid_until')}
              />
            </div>

            <div className="form-group form-group--full">
              <label htmlFor="qf-intro-text">Einleitungstext</label>
              <textarea
                id="qf-intro-text"
                placeholder="Gerne unterbreiten wir Ihnen folgendes Angebot:"
                rows={2}
                {...register('intro_text')}
              />
            </div>

            <div className="form-group form-group--full">
              <label htmlFor="qf-outro-text">Schlusstext</label>
              <textarea
                id="qf-outro-text"
                placeholder="Wir freuen uns auf Ihre Rueckmeldung."
                rows={2}
                {...register('outro_text')}
              />
            </div>

            <div className="form-group form-group--full">
              <label htmlFor="qf-notes">Notizen</label>
              <textarea
                id="qf-notes"
                placeholder="Zusaetzliche Hinweise..."
                rows={2}
                {...register('notes')}
              />
            </div>

            <div className="form-group">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  {...register('kleinunternehmer')}
                />
                <span>Kleinunternehmer gem. &sect;19 UStG</span>
              </label>
            </div>
          </div>

          {/* Quote items */}
          <div className="quote-items-section">
            <h3 className="quote-items-section__title">Positionen</h3>

            <div className="quote-items-header">
              <span className="quote-items-header__col quote-items-header__col--desc">
                Beschreibung
              </span>
              <span className="quote-items-header__col quote-items-header__col--qty">
                Menge
              </span>
              <span className="quote-items-header__col quote-items-header__col--unit">
                Einheit
              </span>
              <span className="quote-items-header__col quote-items-header__col--price">
                Einzelpreis (EUR)
              </span>
              <span className="quote-items-header__col quote-items-header__col--actions" />
            </div>

            {fields.map((field, index) => (
              <div className="quote-item-row" key={field.id}>
                <div className="quote-item-row__col quote-item-row__col--desc">
                  <input
                    type="text"
                    placeholder="Beschreibung"
                    {...register(`items.${index}.description`)}
                  />
                </div>
                <div className="quote-item-row__col quote-item-row__col--qty">
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder="1"
                    {...register(`items.${index}.quantity`)}
                  />
                </div>
                <div className="quote-item-row__col quote-item-row__col--unit">
                  <input
                    type="text"
                    placeholder="Stueck"
                    {...register(`items.${index}.unit`)}
                  />
                </div>
                <div className="quote-item-row__col quote-item-row__col--price">
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder="0.00"
                    {...register(`items.${index}.unit_price`)}
                  />
                </div>
                <div className="quote-item-row__col quote-item-row__col--actions">
                  <button
                    type="button"
                    className="btn-icon btn-icon--danger"
                    onClick={() => {
                      if (fields.length > 1) remove(index);
                    }}
                    disabled={fields.length <= 1}
                    title="Position entfernen"
                  >
                    <X size={16} />
                  </button>
                </div>
              </div>
            ))}

            <button
              type="button"
              className="btn btn--secondary quote-items-section__add"
              onClick={() => append(emptyItem())}
            >
              <Plus size={16} />
              <span>Position hinzufuegen</span>
            </button>

            <div className="quote-total">
              <span className="quote-total__label">Gesamt</span>
              <span className="quote-total__value">
                {formatEurDisplay(totalCents)}
              </span>
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
