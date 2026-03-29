import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { LogOut, User, Mail, Shield, Building2 } from 'lucide-react';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { useAuthStore } from '@/stores/authStore';
import './AccountPage.scss';

export default function AccountPage() {
  const navigate = useNavigate();
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);

  const handleLogout = () => {
    logout();
    navigate('/login', { replace: true });
  };

  if (!user) {
    return (
      <PageWrapper title="Konto">
        <p className="empty-state">Nicht angemeldet.</p>
      </PageWrapper>
    );
  }

  const roleLabels: Record<string, string> = {
    admin: 'Administrator',
    manager: 'Manager',
    user: 'Benutzer',
    viewer: 'Betrachter',
  };

  return (
    <PageWrapper title="Konto">
      <motion.div
        className="account-card"
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <div className="account-card__avatar">
          <User size={40} strokeWidth={1.5} />
        </div>

        <div className="account-info">
          <div className="account-info__row">
            <div className="account-info__icon">
              <User size={18} />
            </div>
            <div className="account-info__content">
              <span className="account-info__label">Benutzer-ID</span>
              <span className="account-info__value account-info__value--mono">
                {user.id}
              </span>
            </div>
          </div>

          <div className="account-info__row">
            <div className="account-info__icon">
              <Mail size={18} />
            </div>
            <div className="account-info__content">
              <span className="account-info__label">E-Mail</span>
              <span className="account-info__value">{user.email}</span>
            </div>
          </div>

          <div className="account-info__row">
            <div className="account-info__icon">
              <Shield size={18} />
            </div>
            <div className="account-info__content">
              <span className="account-info__label">Rolle</span>
              <span className="account-info__value">
                {roleLabels[user.role] || user.role}
              </span>
            </div>
          </div>

          <div className="account-info__row">
            <div className="account-info__icon">
              <Building2 size={18} />
            </div>
            <div className="account-info__content">
              <span className="account-info__label">Mandant-ID</span>
              <span className="account-info__value account-info__value--mono">
                {user.tenant_id}
              </span>
            </div>
          </div>
        </div>

        <div className="account-card__actions">
          <button
            type="button"
            className="btn btn--danger"
            onClick={handleLogout}
          >
            <LogOut size={16} />
            <span>Abmelden</span>
          </button>
        </div>
      </motion.div>
    </PageWrapper>
  );
}
