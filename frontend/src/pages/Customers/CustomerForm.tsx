import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { ArrowLeft, Save } from 'lucide-react';
import toast from 'react-hot-toast';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import * as customerApi from '@/services/customers';
import type { Customer } from '@/types/customer';
import './CustomerForm.scss';

interface CustomerFormData {
  company_name: string;
  customer_number: string;
  email: string;
  phone: string;
  billing_address_street: string;
  billing_address_city: string;
  billing_address_zip: string;
  billing_address_country: string;
  notes: string;
}

const defaultValues: CustomerFormData = {
  company_name: '',
  customer_number: '',
  email: '',
  phone: '',
  billing_address_street: '',
  billing_address_city: '',
  billing_address_zip: '',
  billing_address_country: 'DE',
  notes: '',
};

export default function CustomerForm() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isEdit = !!id;

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<CustomerFormData>({ defaultValues });

  const { data: existing, isLoading } = useQuery({
    queryKey: ['customers', id],
    queryFn: () => customerApi.get(id!),
    enabled: isEdit,
  });

  useEffect(() => {
    if (existing) {
      reset({
        company_name: existing.company_name ?? '',
        customer_number: existing.customer_number ?? '',
        email: existing.email ?? '',
        phone: existing.phone ?? '',
        billing_address_street: existing.billing_address_street ?? '',
        billing_address_city: existing.billing_address_city ?? '',
        billing_address_zip: existing.billing_address_zip ?? '',
        billing_address_country: existing.billing_address_country || 'DE',
        notes: existing.notes ?? '',
      });
    }
  }, [existing, reset]);

  const createMutation = useMutation({
    mutationFn: (data: Partial<Customer>) => customerApi.create(data),
    onSuccess: () => {
      toast.success('Kunde erstellt');
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      navigate('/customers');
    },
    onError: () => toast.error('Fehler beim Erstellen des Kunden'),
  });

  const updateMutation = useMutation({
    mutationFn: (data: Partial<Customer>) => customerApi.update(id!, data),
    onSuccess: () => {
      toast.success('Kunde aktualisiert');
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      queryClient.invalidateQueries({ queryKey: ['customers', id] });
      navigate('/customers');
    },
    onError: () => toast.error('Fehler beim Aktualisieren des Kunden'),
  });

  const onSubmit = (data: CustomerFormData) => {
    if (isEdit) {
      updateMutation.mutate(data);
    } else {
      createMutation.mutate(data);
    }
  };

  const isSaving = createMutation.isPending || updateMutation.isPending;

  if (isEdit && isLoading) {
    return (
      <PageWrapper title="Kunde bearbeiten">
        <div className="detail-skeleton">
          <div className="skeleton skeleton--heading" />
          <div className="skeleton skeleton--block" />
        </div>
      </PageWrapper>
    );
  }

  return (
    <PageWrapper title={isEdit ? 'Kunde bearbeiten' : 'Neuer Kunde'}>
      <button className="btn btn--ghost" onClick={() => navigate('/customers')}>
        <ArrowLeft size={18} />
        <span>Zurueck</span>
      </button>

      <motion.form
        className="customer-form-page"
        onSubmit={handleSubmit(onSubmit)}
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        {/* Stammdaten */}
        <div className="form-card">
          <h2 className="form-section-title">Stammdaten</h2>
          <div className="form-grid">
            <div className="form-group">
              <label htmlFor="company_name">Firma *</label>
              <input
                id="company_name"
                type="text"
                placeholder="Firmenname"
                {...register('company_name', {
                  required: 'Firmenname ist erforderlich',
                })}
              />
              {errors.company_name && (
                <span className="form-error">{errors.company_name.message}</span>
              )}
            </div>

            <div className="form-group">
              <label htmlFor="customer_number">Kundennummer</label>
              <input
                id="customer_number"
                type="text"
                placeholder="z.B. KD-001"
                {...register('customer_number')}
              />
            </div>

            <div className="form-group">
              <label htmlFor="email">E-Mail</label>
              <input
                id="email"
                type="email"
                placeholder="info@firma.de"
                {...register('email', {
                  pattern: {
                    value: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
                    message: 'Ungueltige E-Mail-Adresse',
                  },
                })}
              />
              {errors.email && (
                <span className="form-error">{errors.email.message}</span>
              )}
            </div>

            <div className="form-group">
              <label htmlFor="phone">Telefon</label>
              <input
                id="phone"
                type="text"
                placeholder="+49 123 456789"
                {...register('phone')}
              />
            </div>
          </div>
        </div>

        {/* Rechnungsadresse */}
        <div className="form-card">
          <h2 className="form-section-title">Rechnungsadresse</h2>
          <div className="form-grid">
            <div className="form-group form-group--full">
              <label htmlFor="billing_address_street">Strasse</label>
              <input
                id="billing_address_street"
                type="text"
                placeholder="Musterstrasse 1"
                {...register('billing_address_street')}
              />
            </div>

            <div className="form-group">
              <label htmlFor="billing_address_zip">PLZ</label>
              <input
                id="billing_address_zip"
                type="text"
                placeholder="12345"
                {...register('billing_address_zip')}
              />
            </div>

            <div className="form-group">
              <label htmlFor="billing_address_city">Ort</label>
              <input
                id="billing_address_city"
                type="text"
                placeholder="Musterstadt"
                {...register('billing_address_city')}
              />
            </div>

            <div className="form-group">
              <label htmlFor="billing_address_country">Land</label>
              <input
                id="billing_address_country"
                type="text"
                placeholder="DE"
                {...register('billing_address_country')}
              />
            </div>
          </div>
        </div>

        {/* Notizen */}
        <div className="form-card">
          <h2 className="form-section-title">Notizen</h2>
          <div className="form-grid">
            <div className="form-group form-group--full">
              <label htmlFor="notes">Bemerkungen</label>
              <textarea
                id="notes"
                placeholder="Interne Notizen zum Kunden..."
                {...register('notes')}
              />
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="form-actions">
          <button
            type="button"
            className="btn btn--ghost"
            onClick={() => navigate('/customers')}
          >
            Abbrechen
          </button>
          <button
            type="submit"
            className="btn btn--primary"
            disabled={isSubmitting || isSaving}
          >
            {isSaving ? (
              <span className="loading-spinner loading-spinner--small" />
            ) : (
              <>
                <Save size={16} />
                <span>{isEdit ? 'Speichern' : 'Erstellen'}</span>
              </>
            )}
          </button>
        </div>
      </motion.form>
    </PageWrapper>
  );
}
