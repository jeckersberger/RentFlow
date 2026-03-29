import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useForm, useFieldArray } from 'react-hook-form';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { ArrowLeft, Save, Plus, X } from 'lucide-react';
import toast from 'react-hot-toast';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import * as invoiceApi from '@/services/invoices';
import type { Invoice } from '@/types/invoice';
import './InvoiceForm.scss';

interface InvoiceItemRow {
  description: string;
  quantity: string;
  unit: string;
  unit_price: string;
}

interface InvoiceFormData {
  customer_name: string;
  customer_email: string;
  customer_address: string;
  invoice_date: string;
  due_date: string;
  notes: string;
  kleinunternehmer: boolean;
  items: InvoiceItemRow[];
}

function todayStr(): string {
  return new Date().toISOString().slice(0, 10);
}

function plus14Days(): string {
  const d = new Date();
  d.setDate(d.getDate() + 14);
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

function emptyItem(): InvoiceItemRow {
  return { description: '', quantity: '1', unit: 'Stueck', unit_price: '' };
}

export default function InvoiceForm() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isEdit = !!id;

  const { data: existing, isLoading } = useQuery({
    queryKey: ['invoices', id],
    queryFn: () => invoiceApi.get(id!),
    enabled: isEdit,
  });

  const { data: existingItems } = useQuery({
    queryKey: ['invoices', id, 'items'],
    queryFn: () => invoiceApi.listItems(id!),
    enabled: isEdit,
  });

  const {
    register,
    control,
    handleSubmit,
    reset,
    watch,
    formState: { errors, isDirty },
  } = useForm<InvoiceFormData>({
    defaultValues: {
      customer_name: '',
      customer_email: '',
      customer_address: '',
      invoice_date: todayStr(),
      due_date: plus14Days(),
      notes: '',
      kleinunternehmer: true,
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
      const itemRows: InvoiceItemRow[] =
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
        invoice_date: toDateInput(existing.invoice_date) || todayStr(),
        due_date: toDateInput(existing.due_date) || plus14Days(),
        notes: existing.notes || '',
        kleinunternehmer: existing.kleinunternehmer ?? true,
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
    mutationFn: (payload: Partial<Invoice> & { items?: unknown[] }) =>
      isEdit ? invoiceApi.update(id!, payload) : invoiceApi.create(payload),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['invoices'] });
      toast.success(isEdit ? 'Rechnung aktualisiert' : 'Rechnung erstellt');
      navigate(`/invoices/${result.id}`);
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Fehler beim Speichern');
    },
  });

  const onSubmit = (formData: InvoiceFormData) => {
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
      invoice_date: formData.invoice_date,
      due_date: formData.due_date,
      notes: formData.notes,
      kleinunternehmer: formData.kleinunternehmer,
      items,
    };

    mutation.mutate(payload);
  };

  if (isEdit && isLoading) {
    return (
      <PageWrapper title="Rechnung">
        <div className="detail-skeleton">
          <div className="skeleton skeleton--heading" />
          <div className="skeleton skeleton--block" />
        </div>
      </PageWrapper>
    );
  }

  return (
    <PageWrapper title={isEdit ? 'Rechnung bearbeiten' : 'Neue Rechnung'}>
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
              <label htmlFor="if-customer-name">Kundenname *</label>
              <input
                id="if-customer-name"
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
              <label htmlFor="if-customer-email">E-Mail</label>
              <input
                id="if-customer-email"
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
              <label htmlFor="if-customer-address">Kundenadresse</label>
              <textarea
                id="if-customer-address"
                placeholder="Strasse, PLZ Ort"
                rows={2}
                {...register('customer_address')}
              />
            </div>

            <div className="form-group">
              <label htmlFor="if-invoice-date">Rechnungsdatum</label>
              <input
                id="if-invoice-date"
                type="date"
                {...register('invoice_date')}
              />
            </div>

            <div className="form-group">
              <label htmlFor="if-due-date">Faelligkeitsdatum</label>
              <input
                id="if-due-date"
                type="date"
                {...register('due_date')}
              />
            </div>

            <div className="form-group form-group--full">
              <label htmlFor="if-notes">Notizen</label>
              <textarea
                id="if-notes"
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

          {/* Invoice items */}
          <div className="invoice-items-section">
            <h3 className="invoice-items-section__title">Positionen</h3>

            <div className="invoice-items-header">
              <span className="invoice-items-header__col invoice-items-header__col--desc">
                Beschreibung
              </span>
              <span className="invoice-items-header__col invoice-items-header__col--qty">
                Menge
              </span>
              <span className="invoice-items-header__col invoice-items-header__col--unit">
                Einheit
              </span>
              <span className="invoice-items-header__col invoice-items-header__col--price">
                Einzelpreis (EUR)
              </span>
              <span className="invoice-items-header__col invoice-items-header__col--actions" />
            </div>

            {fields.map((field, index) => (
              <div className="invoice-item-row" key={field.id}>
                <div className="invoice-item-row__col invoice-item-row__col--desc">
                  <input
                    type="text"
                    placeholder="Beschreibung"
                    {...register(`items.${index}.description`)}
                  />
                </div>
                <div className="invoice-item-row__col invoice-item-row__col--qty">
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder="1"
                    {...register(`items.${index}.quantity`)}
                  />
                </div>
                <div className="invoice-item-row__col invoice-item-row__col--unit">
                  <input
                    type="text"
                    placeholder="Stueck"
                    {...register(`items.${index}.unit`)}
                  />
                </div>
                <div className="invoice-item-row__col invoice-item-row__col--price">
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder="0.00"
                    {...register(`items.${index}.unit_price`)}
                  />
                </div>
                <div className="invoice-item-row__col invoice-item-row__col--actions">
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
              className="btn btn--secondary invoice-items-section__add"
              onClick={() => append(emptyItem())}
            >
              <Plus size={16} />
              <span>Position hinzufuegen</span>
            </button>

            <div className="invoice-total">
              <span className="invoice-total__label">Gesamt</span>
              <span className="invoice-total__value">
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
