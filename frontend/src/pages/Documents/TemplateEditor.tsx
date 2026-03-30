import { useState, useRef, useCallback, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { Save, RotateCcw, Eye, FileText } from 'lucide-react';
import toast from 'react-hot-toast';
import { useNavigate } from 'react-router-dom';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import api from '@/services/api';
import type { DocumentTemplate } from '@/types/document';
import './TemplateEditor.scss';

type TemplateType = 'invoice' | 'quote' | 'delivery_note' | 'reminder';

interface TemplateTypeOption {
  value: TemplateType;
  label: string;
}

const TEMPLATE_TYPES: TemplateTypeOption[] = [
  { value: 'invoice', label: 'Rechnung' },
  { value: 'quote', label: 'Angebot' },
  { value: 'delivery_note', label: 'Lieferschein' },
  { value: 'reminder', label: 'Mahnung' },
];

async function fetchTemplateByType(type: TemplateType): Promise<DocumentTemplate | null> {
  const templates = await api.get('/api/v1/document-templates', {
    params: { type },
  }) as unknown as DocumentTemplate[];
  if (Array.isArray(templates) && templates.length > 0) {
    return templates[0];
  }
  return null;
}

async function fetchDefaultPreview(type: TemplateType): Promise<string> {
  const res = await api.get(
    `/api/v1/documents/templates/defaults/${type}/preview`,
  ) as unknown as { content: string } | string;
  if (typeof res === 'string') return res;
  return (res as { content: string }).content ?? '';
}

export default function TemplateEditor() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const iframeRef = useRef<HTMLIFrameElement>(null);

  const [selectedType, setSelectedType] = useState<TemplateType>('invoice');
  const [htmlContent, setHtmlContent] = useState('');
  const [previewHtml, setPreviewHtml] = useState('');
  const [isDirty, setIsDirty] = useState(false);

  const { data: existingTemplate, isLoading: templateLoading } = useQuery({
    queryKey: ['document-template', selectedType],
    queryFn: () => fetchTemplateByType(selectedType),
  });

  const { data: defaultContent, isLoading: defaultLoading } = useQuery({
    queryKey: ['document-template-default', selectedType],
    queryFn: () => fetchDefaultPreview(selectedType),
    enabled: !templateLoading && !existingTemplate,
  });

  // Sync template content into local state when data arrives
  useEffect(() => {
    if (templateLoading) return;
    if (existingTemplate) {
      setHtmlContent(existingTemplate.content);
      setPreviewHtml(existingTemplate.content);
      setIsDirty(false);
    } else if (defaultContent) {
      setHtmlContent(defaultContent);
      setPreviewHtml(defaultContent);
      setIsDirty(false);
    } else {
      setHtmlContent('');
      setPreviewHtml('');
      setIsDirty(false);
    }
  }, [existingTemplate, defaultContent, templateLoading]);

  const saveMutation = useMutation({
    mutationFn: async (content: string) => {
      if (existingTemplate?.id) {
        return api.put(`/api/v1/document-templates/${existingTemplate.id}`, {
          content,
          type: selectedType,
        }) as unknown as DocumentTemplate;
      }
      return api.post('/api/v1/document-templates', {
        name: `${TEMPLATE_TYPES.find((t) => t.value === selectedType)?.label} Template`,
        type: selectedType,
        content,
        is_active: true,
      }) as unknown as DocumentTemplate;
    },
    onSuccess: () => {
      toast.success('Template gespeichert');
      setIsDirty(false);
      queryClient.invalidateQueries({ queryKey: ['document-template', selectedType] });
      queryClient.invalidateQueries({ queryKey: ['document-templates'] });
    },
    onError: () => {
      toast.error('Fehler beim Speichern des Templates');
    },
  });

  const resetMutation = useMutation({
    mutationFn: () => fetchDefaultPreview(selectedType),
    onSuccess: (content: string) => {
      setHtmlContent(content);
      setPreviewHtml(content);
      setIsDirty(true);
      toast.success('Auf Standard zurueckgesetzt');
    },
    onError: () => {
      toast.error('Fehler beim Laden des Standard-Templates');
    },
  });

  const handleTypeChange = useCallback((type: TemplateType) => {
    setSelectedType(type);
    setIsDirty(false);
  }, []);

  const handleContentChange = useCallback((value: string) => {
    setHtmlContent(value);
    setIsDirty(true);
  }, []);

  const handleRefreshPreview = useCallback(() => {
    setPreviewHtml(htmlContent);
  }, [htmlContent]);

  const handleSave = useCallback(() => {
    saveMutation.mutate(htmlContent);
  }, [htmlContent, saveMutation]);

  const handleReset = useCallback(() => {
    resetMutation.mutate();
  }, [resetMutation]);

  const isLoading = templateLoading || defaultLoading;

  return (
    <PageWrapper
      title="Template Editor"
      actions={
        <button
          className="template-editor__back-btn"
          onClick={() => navigate('/documents')}
        >
          Zurueck zu Dokumente
        </button>
      }
    >
      <div className="template-editor__toolbar">
        <div className="template-editor__type-select">
          <label htmlFor="template-type">Dokumenttyp</label>
          <select
            id="template-type"
            value={selectedType}
            onChange={(e) => handleTypeChange(e.target.value as TemplateType)}
            disabled={isLoading}
          >
            {TEMPLATE_TYPES.map((t) => (
              <option key={t.value} value={t.value}>
                {t.label}
              </option>
            ))}
          </select>
        </div>

        <div className="template-editor__actions">
          <button
            className="template-editor__btn template-editor__btn--secondary"
            onClick={handleRefreshPreview}
            disabled={isLoading}
          >
            <Eye size={16} />
            Vorschau aktualisieren
          </button>
          <button
            className="template-editor__btn template-editor__btn--secondary"
            onClick={handleReset}
            disabled={isLoading || resetMutation.isPending}
          >
            <RotateCcw size={16} />
            Auf Standard zuruecksetzen
          </button>
          <button
            className="template-editor__btn template-editor__btn--primary"
            onClick={handleSave}
            disabled={isLoading || saveMutation.isPending || !isDirty}
          >
            <Save size={16} />
            {saveMutation.isPending ? 'Speichert...' : 'Speichern'}
          </button>
        </div>
      </div>

      {isLoading ? (
        <div className="template-editor__loading">
          <div className="loading-spinner" />
          <span>Template wird geladen...</span>
        </div>
      ) : (
        <motion.div
          className="template-editor"
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.2 }}
        >
          <div className="template-editor__panel">
            <div className="template-editor__panel-header">
              <FileText size={16} />
              <span>HTML-Code</span>
              {isDirty && <span className="template-editor__dirty-badge">Ungespeichert</span>}
            </div>
            <textarea
              className="template-editor__code"
              value={htmlContent}
              onChange={(e) => handleContentChange(e.target.value)}
              spellCheck={false}
              placeholder="HTML-Template hier eingeben..."
            />
          </div>

          <div className="template-editor__panel">
            <div className="template-editor__panel-header">
              <Eye size={16} />
              <span>Vorschau</span>
            </div>
            <div className="template-editor__preview-wrapper">
              <iframe
                ref={iframeRef}
                className="template-editor__preview"
                srcDoc={previewHtml}
                title="Template Vorschau"
                sandbox="allow-same-origin"
              />
            </div>
          </div>
        </motion.div>
      )}
    </PageWrapper>
  );
}
