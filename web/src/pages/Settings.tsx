import { useState, useEffect, useCallback } from 'react';
import {
  getProfile,
  getFolders,
  updateFolder,
  updateTags,
  updateBlacklist,
  updateStartTime,
  updateSummaryCount,
  createSchedule,
  deleteSchedule,
  updateProfile,
  deleteAccount,
  type UserProfile,
  ApiError,
} from '../api/client';
import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';
import {
  Loader2,
  Save,
  Trash2,
  Plus,
  X,
  FolderOpen,
  Tag,
  Ban,
  Clock,
  Hash,
  Timer,
  User,
  AlertTriangle,
  Check,
  Play,
  Square,
} from 'lucide-react';

export default function Settings() {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [folders, setFolders] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const { logout } = useAuth();
  const navigate = useNavigate();

  const [selectedFolder, setSelectedFolder] = useState('');
  const [tags, setTags] = useState<string[]>([]);
  const [newTag, setNewTag] = useState('');
  const [blacklist, setBlacklist] = useState<string[]>([]);
  const [newBlacklist, setNewBlacklist] = useState('');
  const [startTime, setStartTime] = useState('');
  const [summaryCount, setSummaryCount] = useState(5);
  const [scheduleInterval, setScheduleInterval] = useState('');
  const [name, setName] = useState('');
  const [receivingEmail, setReceivingEmail] = useState('');

  const showSuccessMsg = (msg: string) => {
    setSuccess(msg);
    setError('');
    setTimeout(() => setSuccess(''), 3000);
  };

  const loadProfile = useCallback(async () => {
    try {
      const data = await getProfile();
      setProfile(data);
      setSelectedFolder(data.folder || '');
      setTags(data.tags || []);
      setBlacklist(data.blackListSenders || []);
      setStartTime(data.startTime ? data.startTime.split('T')[0] : '');
      setSummaryCount(data.summaryCount || 5);
      setScheduleInterval(data.updateInterval || '');
      setName(data.name);
      setReceivingEmail(data.receivingEmail);
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        logout();
        navigate('/login');
      }
    } finally {
      setLoading(false);
    }
  }, [logout, navigate]);

  const loadFolders = useCallback(async () => {
    try {
      const data = await getFolders();
      setFolders(data);
    } catch {
      /* folders fail silently if IMAP not configured */
    }
  }, []);

  useEffect(() => {
    loadProfile();
    loadFolders();
  }, [loadProfile, loadFolders]);

  const handleSave = async (section: string, fn: () => Promise<unknown>) => {
    setSaving(section);
    setError('');
    try {
      await fn();
      showSuccessMsg(`${section} saved`);
      loadProfile();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Save failed');
    } finally {
      setSaving('');
    }
  };

  const handleDelete = async () => {
    try {
      if (profile?.updateInterval && profile.updateInterval !== '0') {
        await deleteSchedule(profile.updateInterval);
      }
      await deleteAccount();
      logout();
      navigate('/login');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Delete failed');
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center py-32 gap-3">
        <Loader2 className="w-8 h-8 animate-spin text-brand-500" />
        <span className="text-sm text-gray-400">Loading settings...</span>
      </div>
    );
  }

  return (
    <div className="max-w-2xl space-y-6 animate-in">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white tracking-tight">Settings</h1>
        <p className="text-gray-500 dark:text-gray-400 mt-1.5">Configure your email summarization preferences.</p>
      </div>

      {/* Toast messages */}
      {error && (
        <div className="bg-red-50 dark:bg-red-950/50 text-red-600 dark:text-red-400 text-sm px-5 py-3.5 rounded-xl border border-red-100 dark:border-red-900/50 animate-in flex items-center gap-2">
          <AlertTriangle className="w-4 h-4 shrink-0" />
          {error}
        </div>
      )}
      {success && (
        <div className="bg-emerald-50 dark:bg-emerald-950/50 text-emerald-600 dark:text-emerald-400 text-sm px-5 py-3.5 rounded-xl border border-emerald-100 dark:border-emerald-900/50 animate-in flex items-center gap-2">
          <Check className="w-4 h-4 shrink-0" />
          {success}
        </div>
      )}

      {/* Profile */}
      <Section title="Profile" icon={<User className="w-5 h-5" />}>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <InputField label="Name" value={name} onChange={setName} placeholder="Jane Doe" />
          <InputField label="Receiving email" value={receivingEmail} onChange={setReceivingEmail} type="email" placeholder="you@gmail.com" />
        </div>
        <SaveButton loading={saving === 'Profile'} onClick={() => handleSave('Profile', () => updateProfile({ name, receivingEmail }))} />
      </Section>

      {/* Folder */}
      <Section title="Email Folder" icon={<FolderOpen className="w-5 h-5" />} description="Select which IMAP folder to scan for emails.">
        <select
          value={selectedFolder}
          onChange={(e) => setSelectedFolder(e.target.value)}
          className="w-full px-4 py-3 border border-gray-200 dark:border-gray-800 rounded-xl bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white focus:ring-2 focus:ring-brand-500/20 focus:border-brand-500 outline-none transition-all duration-200 cursor-pointer"
        >
          <option value="">Select a folder...</option>
          {folders.map((f) => (
            <option key={f} value={f}>{f}</option>
          ))}
        </select>
        {folders.length === 0 && (
          <p className="text-xs text-gray-400 mt-1">Connect your IMAP email to see available folders.</p>
        )}
        <SaveButton loading={saving === 'Folder'} onClick={() => handleSave('Folder', () => updateFolder(selectedFolder))} />
      </Section>

      {/* Tags */}
      <Section title="Email Tags" icon={<Tag className="w-5 h-5" />} description="Emails with these keywords in the subject line will be included in summaries.">
        <div className="flex flex-wrap gap-2 min-h-[2rem]">
          {tags.map((tag) => (
            <span
              key={tag}
              className="inline-flex items-center gap-1.5 px-3.5 py-1.5 bg-brand-50 dark:bg-brand-950/40 text-brand-700 dark:text-brand-300 text-sm font-medium rounded-full border border-brand-200 dark:border-brand-800/50 transition-all hover:shadow-sm"
            >
              {tag}
              <button onClick={() => setTags(tags.filter((t) => t !== tag))} className="hover:text-red-500 transition-colors cursor-pointer">
                <X className="w-3.5 h-3.5" />
              </button>
            </span>
          ))}
          {tags.length === 0 && (
            <span className="text-sm text-gray-400 italic">No tags configured yet.</span>
          )}
        </div>
        <div className="flex gap-2 mt-3">
          <input
            type="text"
            value={newTag}
            onChange={(e) => setNewTag(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && newTag.trim()) {
                e.preventDefault();
                if (!tags.includes(newTag.trim())) {
                  setTags([...tags, newTag.trim()]);
                }
                setNewTag('');
              }
            }}
            placeholder="Type a tag and press Enter..."
            className="flex-1 px-4 py-2.5 border border-gray-200 dark:border-gray-800 rounded-xl bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white placeholder-gray-400 focus:ring-2 focus:ring-brand-500/20 focus:border-brand-500 outline-none transition-all duration-200"
          />
          <button
            onClick={() => {
              if (newTag.trim() && !tags.includes(newTag.trim())) {
                setTags([...tags, newTag.trim()]);
                setNewTag('');
              }
            }}
            className="px-3.5 py-2.5 bg-gray-100 dark:bg-gray-800 hover:bg-brand-50 hover:text-brand-600 dark:hover:bg-brand-950/30 dark:hover:text-brand-400 rounded-xl transition-all duration-200 cursor-pointer"
          >
            <Plus className="w-5 h-5" />
          </button>
        </div>
        <SaveButton loading={saving === 'Tags'} onClick={() => handleSave('Tags', () => updateTags(tags))} />
      </Section>

      {/* Blacklist */}
      <Section title="Blocked Senders" icon={<Ban className="w-5 h-5" />} description="Emails from these addresses will be excluded from summaries.">
        <div className="flex flex-wrap gap-2 min-h-[2rem]">
          {blacklist.map((sender) => (
            <span
              key={sender}
              className="inline-flex items-center gap-1.5 px-3.5 py-1.5 bg-red-50 dark:bg-red-950/30 text-red-600 dark:text-red-400 text-sm font-medium rounded-full border border-red-200 dark:border-red-800/50 transition-all hover:shadow-sm"
            >
              {sender}
              <button onClick={() => setBlacklist(blacklist.filter((s) => s !== sender))} className="hover:text-red-700 transition-colors cursor-pointer">
                <X className="w-3.5 h-3.5" />
              </button>
            </span>
          ))}
          {blacklist.length === 0 && (
            <span className="text-sm text-gray-400 italic">No senders blocked.</span>
          )}
        </div>
        <div className="flex gap-2 mt-3">
          <input
            type="email"
            value={newBlacklist}
            onChange={(e) => setNewBlacklist(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && newBlacklist.trim()) {
                e.preventDefault();
                if (!blacklist.includes(newBlacklist.trim())) {
                  setBlacklist([...blacklist, newBlacklist.trim()]);
                }
                setNewBlacklist('');
              }
            }}
            placeholder="sender@example.com"
            className="flex-1 px-4 py-2.5 border border-gray-200 dark:border-gray-800 rounded-xl bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white placeholder-gray-400 focus:ring-2 focus:ring-brand-500/20 focus:border-brand-500 outline-none transition-all duration-200"
          />
          <button
            onClick={() => {
              if (newBlacklist.trim() && !blacklist.includes(newBlacklist.trim())) {
                setBlacklist([...blacklist, newBlacklist.trim()]);
                setNewBlacklist('');
              }
            }}
            className="px-3.5 py-2.5 bg-gray-100 dark:bg-gray-800 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950/30 dark:hover:text-red-400 rounded-xl transition-all duration-200 cursor-pointer"
          >
            <Plus className="w-5 h-5" />
          </button>
        </div>
        <SaveButton loading={saving === 'Blacklist'} onClick={() => handleSave('Blacklist', () => updateBlacklist(blacklist))} />
      </Section>

      {/* Start Time */}
      <Section title="Start Date" icon={<Clock className="w-5 h-5" />} description="Only process emails sent after this date.">
        <input
          type="date"
          value={startTime}
          onChange={(e) => setStartTime(e.target.value)}
          className="w-full px-4 py-3 border border-gray-200 dark:border-gray-800 rounded-xl bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white focus:ring-2 focus:ring-brand-500/20 focus:border-brand-500 outline-none transition-all duration-200"
        />
        <SaveButton loading={saving === 'StartTime'} onClick={() => handleSave('StartTime', () => updateStartTime(startTime ? `${startTime}T00:00:00Z` : ''))} />
      </Section>

      {/* Summary Count */}
      <Section title="Summary Length" icon={<Hash className="w-5 h-5" />} description="Number of key sentences extracted per summary.">
        <div className="flex items-center gap-4">
          <input
            type="range"
            value={summaryCount}
            onChange={(e) => setSummaryCount(Number(e.target.value))}
            min={1}
            max={20}
            className="flex-1 accent-brand-500"
          />
          <span className="w-12 text-center text-lg font-bold text-brand-600 dark:text-brand-400 tabular-nums">{summaryCount}</span>
        </div>
        <SaveButton loading={saving === 'SummaryCount'} onClick={() => handleSave('SummaryCount', () => updateSummaryCount(summaryCount))} />
      </Section>

      {/* Schedule */}
      <Section title="Automatic Schedule" icon={<Timer className="w-5 h-5" />} description="Automatically generate and email summaries at a set interval.">
        {profile?.updateInterval && profile.updateInterval !== '0' ? (
          <div className="flex items-center justify-between bg-emerald-50 dark:bg-emerald-950/30 rounded-xl px-5 py-4 border border-emerald-200 dark:border-emerald-800/50">
            <div className="flex items-center gap-3">
              <div className="w-2.5 h-2.5 rounded-full bg-emerald-500 animate-pulse" />
              <span className="text-sm font-medium text-emerald-700 dark:text-emerald-300">
                Running every <strong>{profile.updateInterval}</strong> minutes
              </span>
            </div>
            <button
              onClick={() =>
                handleSave('Schedule', async () => {
                  await deleteSchedule(profile.updateInterval);
                  setScheduleInterval('');
                })
              }
              className="flex items-center gap-1.5 text-red-500 hover:text-red-600 dark:hover:text-red-400 text-sm font-semibold transition-colors cursor-pointer"
            >
              {saving === 'Schedule' ? <Loader2 className="w-4 h-4 animate-spin" /> : <Square className="w-4 h-4" />}
              Stop
            </button>
          </div>
        ) : (
          <div className="flex gap-3">
            <div className="relative flex-1">
              <input
                type="number"
                value={scheduleInterval}
                onChange={(e) => setScheduleInterval(e.target.value)}
                min={1}
                placeholder="Minutes between summaries"
                className="w-full px-4 py-3 border border-gray-200 dark:border-gray-800 rounded-xl bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white placeholder-gray-400 focus:ring-2 focus:ring-brand-500/20 focus:border-brand-500 outline-none transition-all duration-200"
              />
            </div>
            <button
              onClick={() => handleSave('Schedule', () => createSchedule(scheduleInterval))}
              disabled={!scheduleInterval || saving === 'Schedule'}
              className="flex items-center gap-2 px-5 py-3 gradient-brand hover:opacity-90 disabled:opacity-40 text-white font-semibold rounded-xl transition-all duration-200 shadow-lg shadow-brand-500/25 cursor-pointer"
            >
              {saving === 'Schedule' ? <Loader2 className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
              Start
            </button>
          </div>
        )}
      </Section>

      {/* Danger Zone */}
      <div className="rounded-2xl border-2 border-dashed border-red-200 dark:border-red-800/50 p-6 sm:p-8 animate-in-delay-2">
        <div className="flex items-center gap-2.5 mb-2">
          <AlertTriangle className="w-5 h-5 text-red-500" />
          <h2 className="text-lg font-semibold text-red-600 dark:text-red-400">Danger Zone</h2>
        </div>
        <p className="text-sm text-gray-500 dark:text-gray-400 mb-5">
          Permanently delete your account and all associated data. This action cannot be undone.
        </p>
        {showDeleteConfirm ? (
          <div className="flex items-center gap-3 animate-in">
            <button
              onClick={handleDelete}
              className="px-5 py-2.5 bg-red-600 hover:bg-red-700 text-white text-sm font-semibold rounded-xl transition-all shadow-lg shadow-red-500/25 cursor-pointer"
            >
              Yes, delete everything
            </button>
            <button
              onClick={() => setShowDeleteConfirm(false)}
              className="px-5 py-2.5 text-sm text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 font-medium transition-colors cursor-pointer"
            >
              Cancel
            </button>
          </div>
        ) : (
          <button
            onClick={() => setShowDeleteConfirm(true)}
            className="flex items-center gap-2 px-5 py-2.5 border border-red-200 dark:border-red-800/50 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-950/30 text-sm font-semibold rounded-xl transition-all duration-200 cursor-pointer"
          >
            <Trash2 className="w-4 h-4" />
            Delete Account
          </button>
        )}
      </div>
    </div>
  );
}

function Section({
  title,
  icon,
  description,
  children,
}: {
  title: string;
  icon: React.ReactNode;
  description?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-800 p-6 sm:p-8 space-y-4 transition-all duration-200 hover:shadow-md">
      <div>
        <div className="flex items-center gap-2.5">
          <span className="text-brand-500 dark:text-brand-400">{icon}</span>
          <h2 className="text-base font-semibold text-gray-900 dark:text-white">{title}</h2>
        </div>
        {description && (
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1 ml-[30px]">{description}</p>
        )}
      </div>
      {children}
    </div>
  );
}

function InputField({
  label,
  value,
  onChange,
  type = 'text',
  placeholder,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  type?: string;
  placeholder?: string;
}) {
  return (
    <div className="space-y-1.5">
      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">{label}</label>
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full px-4 py-3 border border-gray-200 dark:border-gray-800 rounded-xl bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white placeholder-gray-400 focus:ring-2 focus:ring-brand-500/20 focus:border-brand-500 outline-none transition-all duration-200"
      />
    </div>
  );
}

function SaveButton({ loading, onClick }: { loading: boolean; onClick: () => void }) {
  return (
    <div className="flex justify-end pt-1">
      <button
        onClick={onClick}
        disabled={loading}
        className="flex items-center gap-2 px-5 py-2.5 gradient-brand hover:opacity-90 disabled:opacity-50 text-white text-sm font-semibold rounded-xl transition-all duration-200 shadow-md shadow-brand-500/20 cursor-pointer"
      >
        {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
        Save
      </button>
    </div>
  );
}
