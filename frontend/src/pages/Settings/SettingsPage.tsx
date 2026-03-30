import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Building2,
  Coins,
  Users,
  Mail,
  Server,
  Save,
  UserPlus,
  RefreshCw,
} from 'lucide-react';
import toast from 'react-hot-toast';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import api from '@/services/api';
import './SettingsPage.scss';

// ── Types ──

type TabKey = 'firma' | 'finanzen' | 'benutzer' | 'email' | 'system';

interface DunningConfig {
  reminder_days: number;
  dunning1_days: number;
  dunning2_days: number;
  reminder_fee: number;
  dunning1_fee: number;
  dunning2_fee: number;
  auto_send: boolean;
}

interface UserRow {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  role: string;
  is_active: boolean;
}

interface EmailStatus {
  configured: boolean;
  provider?: string;
}

interface AiStatus {
  configured: boolean;
  model?: string;
}

const tabs: { key: TabKey; label: string; icon: typeof Building2 }[] = [
  { key: 'firma', label: 'Firma', icon: Building2 },
  { key: 'finanzen', label: 'Finanzen', icon: Coins },
  { key: 'benutzer', label: 'Benutzer', icon: Users },
  { key: 'email', label: 'E-Mail', icon: Mail },
  { key: 'system', label: 'System', icon: Server },
];

const defaultDunning: DunningConfig = {
  reminder_days: 14,
  dunning1_days: 28,
  dunning2_days: 42,
  reminder_fee: 0,
  dunning1_fee: 500,
  dunning2_fee: 1000,
  auto_send: false,
};

// ── Helpers ──

function FieldGroup({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="settings-field">
      <label>{label}</label>
      {children}
    </div>
  );
}

// ── Component ──

export default function SettingsPage() {
  const queryClient = useQueryClient();
  const [tab, setTab] = useState<TabKey>('firma');

  // ── Dunning config ──
  const { data: dunningRaw } = useQuery({
    queryKey: ['dunning-config'],
    queryFn: () => api.get('/api/v1/dunning/config') as unknown as DunningConfig,
  });
  const dunningConfig = dunningRaw ?? defaultDunning;

  const [dunningForm, setDunningForm] = useState<DunningConfig | null>(null);
  const activeDunning = dunningForm ?? dunningConfig;

  const dunningMutation = useMutation({
    mutationFn: async (cfg: DunningConfig) => {
      const result = await api.put('/api/v1/dunning/config', cfg);
      return result as unknown as DunningConfig;
    },
    onSuccess: () => {
      toast.success('Mahnwesen-Konfiguration gespeichert');
      queryClient.invalidateQueries({ queryKey: ['dunning-config'] });
      setDunningForm(null);
    },
    onError: () => toast.error('Fehler beim Speichern'),
  });

  function updateDunning<K extends keyof DunningConfig>(
    key: K,
    value: DunningConfig[K],
  ) {
    setDunningForm((prev) => ({ ...(prev ?? dunningConfig), [key]: value }));
  }

  // ── Users ──
  const { data: usersRaw, isLoading: usersLoading } = useQuery({
    queryKey: ['settings-users'],
    queryFn: () => api.get('/api/v1/users') as unknown as UserRow[],
    enabled: tab === 'benutzer',
  });
  const users: UserRow[] = Array.isArray(usersRaw) ? usersRaw : [];

  const userColumns: Column<UserRow>[] = [
    {
      key: 'name',
      label: 'Name',
      render: (r) => `${r.first_name ?? ''} ${r.last_name ?? ''}`.trim() || '-',
    },
    { key: 'email', label: 'E-Mail' },
    {
      key: 'role',
      label: 'Rolle',
      render: (r) => (
        <span className="settings-role-badge">{r.role ?? 'user'}</span>
      ),
    },
    {
      key: 'is_active',
      label: 'Status',
      render: (r) => (
        <span
          className={`settings-status-pill ${r.is_active ? 'settings-status-pill--active' : 'settings-status-pill--inactive'}`}
        >
          {r.is_active ? 'Aktiv' : 'Inaktiv'}
        </span>
      ),
    },
  ];

  // ── Email status ──
  const { data: emailStatus } = useQuery({
    queryKey: ['email-status'],
    queryFn: () => api.get('/api/v1/email/status') as unknown as EmailStatus,
    enabled: tab === 'email',
  });

  // ── AI status ──
  const { data: aiStatus } = useQuery({
    queryKey: ['ai-status'],
    queryFn: () => api.get('/api/v1/ai/status') as unknown as AiStatus,
    enabled: tab === 'system',
  });

  // ── Render ──

  return (
    <PageWrapper title="Einstellungen">
      {/* Tab bar */}
      <div className="settings-tabs">
        {tabs.map((t) => {
          const Icon = t.icon;
          return (
            <button
              key={t.key}
              className={`settings-tab ${tab === t.key ? 'settings-tab--active' : ''}`}
              onClick={() => setTab(t.key)}
            >
              <Icon size={16} />
              <span>{t.label}</span>
            </button>
          );
        })}
      </div>

      {/* Tab content */}
      <AnimatePresence mode="wait">
        <motion.div
          key={tab}
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -4 }}
          transition={{ duration: 0.2 }}
        >
          {/* ───── Tab 1: Firma ───── */}
          {tab === 'firma' && (
            <div className="settings-card">
              <h3>Firmendaten</h3>
              <p className="settings-hint" style={{ marginBottom: 20 }}>
                Wird in Mandanten-Einstellungen verwaltet. Aenderungen wirken
                sich auf alle generierten Dokumente aus.
              </p>
              <div className="settings-grid settings-grid--3">
                {[
                  ['Firmenname', 'JE Sound & Light GmbH'],
                  ['Strasse', 'Musterstrasse 12'],
                  ['PLZ', '4020'],
                  ['Ort', 'Linz'],
                  ['Telefon', '+43 660 1234567'],
                  ['E-Mail', 'office@example.at'],
                  ['Steuernummer', 'ATU12345678'],
                  ['IBAN', 'AT12 3456 7890 1234 5678'],
                  ['BIC', 'BKAUATWW'],
                  ['Bankname', 'Erste Bank'],
                ].map(([label, placeholder]) => (
                  <FieldGroup key={label} label={label}>
                    <input
                      type="text"
                      disabled
                      placeholder={placeholder}
                      value=""
                      readOnly
                    />
                  </FieldGroup>
                ))}
              </div>
              <p className="settings-hint" style={{ marginTop: 16 }}>
                Diese Felder werden zentral verwaltet und koennen hier nicht
                bearbeitet werden.
              </p>
            </div>
          )}

          {/* ───── Tab 2: Finanzen ───── */}
          {tab === 'finanzen' && (
            <div className="settings-card">
              <div className="settings-card__header">
                <h3>Mahnwesen-Konfiguration</h3>
                <button
                  className="btn btn--primary"
                  disabled={!dunningForm || dunningMutation.isPending}
                  onClick={() => {
                    if (dunningForm) dunningMutation.mutate(dunningForm);
                  }}
                >
                  <Save size={16} />
                  {dunningMutation.isPending ? 'Speichern...' : 'Speichern'}
                </button>
              </div>

              <div className="settings-grid">
                <FieldGroup label="Zahlungserinnerung nach (Tage)">
                  <input
                    type="number"
                    min={1}
                    value={activeDunning.reminder_days}
                    onChange={(e) =>
                      updateDunning('reminder_days', Number(e.target.value))
                    }
                  />
                </FieldGroup>
                <FieldGroup label="1. Mahnung nach (Tage)">
                  <input
                    type="number"
                    min={1}
                    value={activeDunning.dunning1_days}
                    onChange={(e) =>
                      updateDunning('dunning1_days', Number(e.target.value))
                    }
                  />
                </FieldGroup>
                <FieldGroup label="2. Mahnung nach (Tage)">
                  <input
                    type="number"
                    min={1}
                    value={activeDunning.dunning2_days}
                    onChange={(e) =>
                      updateDunning('dunning2_days', Number(e.target.value))
                    }
                  />
                </FieldGroup>
                <FieldGroup label="Erinnerungsgebuehr (EUR)">
                  <input
                    type="number"
                    min={0}
                    step={0.01}
                    value={(activeDunning.reminder_fee / 100).toFixed(2)}
                    onChange={(e) =>
                      updateDunning(
                        'reminder_fee',
                        Math.round(Number(e.target.value) * 100),
                      )
                    }
                  />
                </FieldGroup>
                <FieldGroup label="Gebuehr 1. Mahnung (EUR)">
                  <input
                    type="number"
                    min={0}
                    step={0.01}
                    value={(activeDunning.dunning1_fee / 100).toFixed(2)}
                    onChange={(e) =>
                      updateDunning(
                        'dunning1_fee',
                        Math.round(Number(e.target.value) * 100),
                      )
                    }
                  />
                </FieldGroup>
                <FieldGroup label="Gebuehr 2. Mahnung (EUR)">
                  <input
                    type="number"
                    min={0}
                    step={0.01}
                    value={(activeDunning.dunning2_fee / 100).toFixed(2)}
                    onChange={(e) =>
                      updateDunning(
                        'dunning2_fee',
                        Math.round(Number(e.target.value) * 100),
                      )
                    }
                  />
                </FieldGroup>
              </div>

              <div className="settings-toggle-row">
                <label className="settings-toggle">
                  <input
                    type="checkbox"
                    checked={activeDunning.auto_send}
                    onChange={(e) =>
                      updateDunning('auto_send', e.target.checked)
                    }
                  />
                  <span className="settings-toggle__slider" />
                </label>
                <span>Mahnungen automatisch versenden</span>
              </div>
            </div>
          )}

          {/* ───── Tab 3: Benutzer ───── */}
          {tab === 'benutzer' && (
            <div className="settings-card">
              <div className="settings-card__header">
                <h3>Benutzer</h3>
                <button
                  className="btn btn--primary"
                  onClick={() =>
                    toast('Einladungssystem wird eingerichtet', {
                      icon: '\u{2709}\uFE0F',
                    })
                  }
                >
                  <UserPlus size={16} />
                  Einladen
                </button>
              </div>
              <DataTable
                columns={userColumns}
                data={users}
                loading={usersLoading}
                emptyMessage="Keine Benutzer gefunden."
              />
            </div>
          )}

          {/* ───── Tab 4: E-Mail ───── */}
          {tab === 'email' && (
            <div className="settings-card">
              <h3>E-Mail (SMTP)</h3>
              <div className="settings-status-row">
                <span
                  className={`settings-dot ${emailStatus?.configured ? 'settings-dot--ok' : 'settings-dot--error'}`}
                />
                <span className="settings-status-label">
                  {emailStatus?.configured
                    ? 'SMTP konfiguriert'
                    : 'SMTP nicht konfiguriert'}
                </span>
              </div>
              {emailStatus?.configured && emailStatus.provider && (
                <p className="settings-hint">
                  Provider: {emailStatus.provider}
                </p>
              )}
              {!emailStatus?.configured && (
                <div className="settings-instructions">
                  <p>
                    Um den E-Mail-Versand zu aktivieren, muessen folgende
                    Umgebungsvariablen auf dem Server gesetzt werden:
                  </p>
                  <ul>
                    <li>
                      <code>SMTP_HOST</code> &mdash; z.B.{' '}
                      <code>smtp.mailgun.org</code>
                    </li>
                    <li>
                      <code>SMTP_PORT</code> &mdash; z.B. <code>587</code>
                    </li>
                    <li>
                      <code>SMTP_USER</code> &mdash; Benutzername
                    </li>
                    <li>
                      <code>SMTP_PASSWORD</code> &mdash; Passwort
                    </li>
                    <li>
                      <code>SMTP_FROM</code> &mdash; Absenderadresse
                    </li>
                  </ul>
                  <p className="settings-hint">
                    Nach dem Setzen der Variablen den Notification-Service neu
                    starten.
                  </p>
                </div>
              )}
            </div>
          )}

          {/* ───── Tab 5: System ───── */}
          {tab === 'system' && (
            <div className="settings-card">
              <h3>System</h3>
              <div className="settings-grid">
                <FieldGroup label="App-Version">
                  <span className="settings-value">v1.1.0</span>
                </FieldGroup>
                <FieldGroup label="KI-Status">
                  <div className="settings-status-row">
                    <span
                      className={`settings-dot ${aiStatus?.configured ? 'settings-dot--ok' : 'settings-dot--error'}`}
                    />
                    <span>
                      {aiStatus?.configured
                        ? aiStatus.model ?? 'Aktiv'
                        : 'Nicht konfiguriert'}
                    </span>
                  </div>
                </FieldGroup>
              </div>

              <div className="settings-actions-row">
                <button
                  className="btn btn--secondary"
                  onClick={() =>
                    toast.success('Keine Updates verfuegbar. Sie nutzen die aktuelle Version.')
                  }
                >
                  <RefreshCw size={16} />
                  Nach Updates suchen
                </button>
              </div>
            </div>
          )}
        </motion.div>
      </AnimatePresence>
    </PageWrapper>
  );
}
