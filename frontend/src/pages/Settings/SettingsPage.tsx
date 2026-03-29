import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { Save } from 'lucide-react';
import toast from 'react-hot-toast';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import api from '@/services/api';
import './SettingsPage.scss';

interface DunningConfig {
  reminder_days: number;
  dunning1_days: number;
  dunning2_days: number;
  reminder_fee: number;
  dunning1_fee: number;
  dunning2_fee: number;
  auto_send: boolean;
}

export default function SettingsPage() {
  const queryClient = useQueryClient();
  const [tab, setTab] = useState<'dunning' | 'email' | 'company'>('company');

  const { data: dunningConfig } = useQuery({
    queryKey: ['dunning-config'],
    queryFn: () => api.get('/api/v1/dunning/config') as unknown as DunningConfig,
  });

  const { data: emailStatus } = useQuery({
    queryKey: ['email-status'],
    queryFn: () => api.get('/api/v1/email/status') as unknown as { configured: boolean },
  });

  return (
    <PageWrapper title="Einstellungen">
      <div className="settings-tabs">
        <button className={`settings-tab ${tab === 'company' ? 'settings-tab--active' : ''}`} onClick={() => setTab('company')}>
          Firma
        </button>
        <button className={`settings-tab ${tab === 'dunning' ? 'settings-tab--active' : ''}`} onClick={() => setTab('dunning')}>
          Mahnwesen
        </button>
        <button className={`settings-tab ${tab === 'email' ? 'settings-tab--active' : ''}`} onClick={() => setTab('email')}>
          E-Mail
        </button>
      </div>

      <motion.div
        key={tab}
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.2 }}
      >
        {tab === 'company' && (
          <div className="settings-card">
            <h3>Firmendaten</h3>
            <p className="settings-hint">
              Firmendaten werden in den Einstellungen des Mandanten verwaltet.
              Aenderungen wirken sich auf alle generierten Dokumente aus.
            </p>
          </div>
        )}

        {tab === 'dunning' && dunningConfig && (
          <div className="settings-card">
            <h3>Mahnwesen-Konfiguration</h3>
            <div className="settings-grid">
              <div className="settings-field">
                <label>Zahlungserinnerung nach (Tage)</label>
                <span className="settings-value">{dunningConfig.reminder_days}</span>
              </div>
              <div className="settings-field">
                <label>1. Mahnung nach (Tage)</label>
                <span className="settings-value">{dunningConfig.dunning1_days}</span>
              </div>
              <div className="settings-field">
                <label>2. Mahnung nach (Tage)</label>
                <span className="settings-value">{dunningConfig.dunning2_days}</span>
              </div>
              <div className="settings-field">
                <label>Mahngebuehr 1. Mahnung</label>
                <span className="settings-value">{(dunningConfig.dunning1_fee / 100).toFixed(2)} EUR</span>
              </div>
              <div className="settings-field">
                <label>Mahngebuehr 2. Mahnung</label>
                <span className="settings-value">{(dunningConfig.dunning2_fee / 100).toFixed(2)} EUR</span>
              </div>
              <div className="settings-field">
                <label>Automatisch senden</label>
                <span className="settings-value">{dunningConfig.auto_send ? 'Ja' : 'Nein'}</span>
              </div>
            </div>
          </div>
        )}

        {tab === 'email' && (
          <div className="settings-card">
            <h3>E-Mail (SMTP)</h3>
            <div className="settings-status">
              <span className={`settings-dot ${emailStatus?.configured ? 'settings-dot--ok' : 'settings-dot--error'}`} />
              <span>{emailStatus?.configured ? 'SMTP konfiguriert' : 'SMTP nicht konfiguriert'}</span>
            </div>
            {!emailStatus?.configured && (
              <p className="settings-hint">
                Setze SMTP_HOST, SMTP_USER und SMTP_PASSWORD in der Server-Konfiguration (.env).
              </p>
            )}
          </div>
        )}
      </motion.div>
    </PageWrapper>
  );
}
