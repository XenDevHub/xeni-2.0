'use client';

import { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { Link } from '@/i18n/routing';
import { Store, Save, Truck, Smartphone, Brain, Lock, Info } from 'lucide-react';
import api from '@/lib/api';
import toast from 'react-hot-toast';

interface Shop {
  id: string;
  shop_name: string;
  shop_description: string | null;
  shop_logo_url: string | null;
  shop_banner_url?: string | null;
  sslcommerz_store_id?: string | null;
  sslcommerz_is_live?: boolean;
  preferred_language: string;
  courier_preference: string;
  bkash_merchant_number: string | null;
  nagad_merchant_number: string | null;
  owner_mobile?: string | null;
  district?: string | null;
  delivery_charge_inside?: number;
  delivery_charge_outside?: number;
  payment_verification_mode?: string;
  bkash_app_key?: string | null;
  nagad_merchant_id?: string | null;
  auto_reply_enabled: boolean;
  auto_order_enabled: boolean;
  custom_agent_rules?: string | null;
  integrations?: Record<string, string>;
}

export default function ShopPage() {
  const [shop, setShop] = useState<Shop | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [isNew, setIsNew] = useState(false);
  const [globalRules, setGlobalRules] = useState('');
  const [form, setForm] = useState({
    shop_name: '',
    shop_description: '',
    shop_logo_url: '',
    shop_banner_url: '',
    preferred_language: 'bn',
    courier_preference: 'pathao',
    bkash_merchant_number: '',
    nagad_merchant_number: '',
    owner_mobile: '',
    district: '',
    delivery_charge_inside: 60,
    delivery_charge_outside: 120,
    payment_verification_mode: 'manual',
    bkash_app_key: '',
    bkash_app_secret: '',
    nagad_merchant_id: '',
    nagad_merchant_key: '',
    sslcommerz_store_id: '',
    sslcommerz_store_passwd: '',
    sslcommerz_is_live: false,
    auto_reply_enabled: true,
    auto_order_enabled: true,
    custom_agent_rules: '',
    integrations: {
      pathao_client_id: '',
      pathao_client_secret: '',
      steadfast_api_key: '',
      steadfast_secret_key: ''
    }
  });

  useEffect(() => {
    fetchShopAndRules();
  }, []);

  async function fetchShopAndRules() {
    try {
      const [shopRes, rulesRes] = await Promise.allSettled([
        api.get('/shops/me'),
        api.get('/user/global-rules'),
      ]);

      if (shopRes.status === 'fulfilled') {
        const s = shopRes.value.data.data;
        setShop(s);
        setForm({
          shop_name: s.shop_name || '',
          shop_description: s.shop_description || '',
          shop_logo_url: s.shop_logo_url || '',
          shop_banner_url: s.shop_banner_url || '',
          preferred_language: s.preferred_language || 'bn',
          courier_preference: s.courier_preference || 'pathao',
          bkash_merchant_number: s.bkash_merchant_number || '',
          nagad_merchant_number: s.nagad_merchant_number || '',
          owner_mobile: s.owner_mobile || '',
          district: s.district || '',
          delivery_charge_inside: s.delivery_charge_inside ?? 60,
          delivery_charge_outside: s.delivery_charge_outside ?? 120,
          payment_verification_mode: s.payment_verification_mode || 'manual',
          bkash_app_key: s.bkash_app_key || '',
          bkash_app_secret: '', // Security: don't show secret
          nagad_merchant_id: s.nagad_merchant_id || '',
          nagad_merchant_key: '', // Security: don't show secret
          sslcommerz_store_id: s.sslcommerz_store_id || '',
          sslcommerz_store_passwd: '',
          sslcommerz_is_live: s.sslcommerz_is_live ?? false,
          auto_reply_enabled: s.auto_reply_enabled ?? true,
          auto_order_enabled: s.auto_order_enabled ?? true,
          custom_agent_rules: s.custom_agent_rules || '',
          integrations: {
            pathao_client_id: s.integrations?.pathao_client_id || '',
            pathao_client_secret: s.integrations?.pathao_client_secret || '',
            steadfast_api_key: s.integrations?.steadfast_api_key || '',
            steadfast_secret_key: s.integrations?.steadfast_secret_key || ''
          }
        });
      } else {
        setIsNew(true);
      }

      if (rulesRes.status === 'fulfilled') {
        setGlobalRules(rulesRes.value.data.data?.setting_value || '');
      }
    } catch {
      setIsNew(true);
    }
    setLoading(false);
  }

  const handleSave = async () => {
    if (!form.shop_name.trim()) {
      toast.error('Shop name is required');
      return;
    }
    setSaving(true);
    try {
      if (isNew) {
        const res = await api.post('/shops', form);
        setShop(res.data.data);
        setIsNew(false);
        toast.success('Shop created! 🎉');
      } else {
        const res = await api.put('/shops/me', form);
        setShop(res.data.data);
        toast.success('Shop updated!');
      }
    } catch (err: any) {
      toast.error(err.response?.data?.error || 'Failed to save');
    }
    setSaving(false);
  };

  if (loading) {
    return <div className="p-8"><div className="skeleton w-full h-96" /></div>;
  }

  return (
    <div className="p-6 lg:p-8 max-w-3xl">
      <div className="mb-8">
        <h1 className="text-2xl font-heading font-bold flex items-center gap-3" style={{ color: 'var(--text-primary)' }}>
          <Store className="w-7 h-7 text-primary" />
          {isNew ? 'Set Up Your Shop' : 'Shop Settings'}
        </h1>
        <p className="mt-1" style={{ color: 'var(--text-muted)' }}>
          {isNew ? 'Create your shop to start selling on Messenger.' : 'Manage your shop configuration.'}
        </p>
      </div>

      <div className="space-y-6">
        {/* ── Main Shop Form ── */}
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} className="glass-card p-6 space-y-6">
          {/* Shop Name */}
          <div>
            <label className="block text-sm font-medium mb-2" style={{ color: 'var(--text-secondary)' }}>Shop Name *</label>
            <input
              className="input-field"
              placeholder="e.g., Trendy Fashion BD"
              value={form.shop_name}
              onChange={e => setForm({ ...form, shop_name: e.target.value })}
            />
          </div>

          {/* Description */}
          <div>
            <label className="block text-sm font-medium mb-2" style={{ color: 'var(--text-secondary)' }}>Description</label>
            <textarea
              className="input-field min-h-[100px] resize-none"
              placeholder="Describe your shop..."
              value={form.shop_description}
              onChange={e => setForm({ ...form, shop_description: e.target.value })}
            />
          </div>

          {/* Shop Branding (Logo & Banner) */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2" style={{ color: 'var(--text-secondary)' }}>Shop Logo URL</label>
              <input
                className="input-field"
                placeholder="https://.../logo.png"
                value={form.shop_logo_url}
                onChange={e => setForm({ ...form, shop_logo_url: e.target.value })}
              />
              {form.shop_logo_url && (
                <div className="mt-2 w-12 h-12 rounded-xl overflow-hidden border border-gray-600">
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img src={form.shop_logo_url} alt="Logo preview" className="w-full h-full object-cover" />
                </div>
              )}
            </div>
            <div>
              <label className="block text-sm font-medium mb-2" style={{ color: 'var(--text-secondary)' }}>Shop Banner URL (E-Pic Storefront)</label>
              <input
                className="input-field"
                placeholder="https://.../banner.jpg (16:9 or 4:1 banner)"
                value={form.shop_banner_url}
                onChange={e => setForm({ ...form, shop_banner_url: e.target.value })}
              />
              {form.shop_banner_url && (
                <div className="mt-2 aspect-[4/1] rounded-xl overflow-hidden border border-gray-600">
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img src={form.shop_banner_url} alt="Banner preview" className="w-full h-full object-cover" />
                </div>
              )}
            </div>
          </div>

          {/* Contact & Location */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2 flex items-center gap-2" style={{ color: 'var(--text-secondary)' }}>
                <Smartphone className="w-4 h-4" /> Owner Mobile (WhatsApp)
              </label>
              <input 
                className="input-field" 
                placeholder="01XXXXXXXXX" 
                value={form.owner_mobile} 
                onChange={e => setForm({ ...form, owner_mobile: e.target.value })} 
                title="Used for manual payment review notifications"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2" style={{ color: 'var(--text-secondary)' }}>Shop District</label>
              <input 
                className="input-field" 
                placeholder="e.g. Dhaka" 
                value={form.district} 
                onChange={e => setForm({ ...form, district: e.target.value })} 
              />
            </div>
          </div>

          {/* Language & Courier */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2" style={{ color: 'var(--text-secondary)' }}>Language</label>
              <select className="input-field" value={form.preferred_language} onChange={e => setForm({ ...form, preferred_language: e.target.value })}>
                <option value="bn">বাংলা (Bangla)</option>
                <option value="en">English</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-2 flex items-center gap-2" style={{ color: 'var(--text-secondary)' }}>
                <Truck className="w-4 h-4" /> Courier
              </label>
              <select className="input-field" value={form.courier_preference} onChange={e => setForm({ ...form, courier_preference: e.target.value })}>
                <option value="pathao">Pathao</option>
                <option value="steadfast">Steadfast</option>
              </select>
            </div>
          </div>

          {/* Delivery Charges */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2" style={{ color: 'var(--text-secondary)' }}>Delivery Charge (Inside District)</label>
              <div className="relative">
                <span className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400">৳</span>
                <input type="number" min="0" className="input-field pl-8" value={form.delivery_charge_inside} onChange={e => setForm({ ...form, delivery_charge_inside: Number(e.target.value) })} />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium mb-2" style={{ color: 'var(--text-secondary)' }}>Delivery Charge (Outside District)</label>
              <div className="relative">
                <span className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400">৳</span>
                <input type="number" min="0" className="input-field pl-8" value={form.delivery_charge_outside} onChange={e => setForm({ ...form, delivery_charge_outside: Number(e.target.value) })} />
              </div>
            </div>
          </div>

          {/* MFS Numbers */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2 flex items-center gap-2" style={{ color: 'var(--text-secondary)' }}>
                <Smartphone className="w-4 h-4" /> bKash Number
              </label>
              <input className="input-field" placeholder="01XXXXXXXXX" value={form.bkash_merchant_number} onChange={e => setForm({ ...form, bkash_merchant_number: e.target.value })} />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2 flex items-center gap-2" style={{ color: 'var(--text-secondary)' }}>
                <Smartphone className="w-4 h-4" /> Nagad Number
              </label>
              <input className="input-field" placeholder="01XXXXXXXXX" value={form.nagad_merchant_number} onChange={e => setForm({ ...form, nagad_merchant_number: e.target.value })} />
            </div>
          </div>

          {/* Toggles */}
          <div className="space-y-4 pt-2 border-t" style={{ borderColor: 'var(--border-color)' }}>
            <p className="text-xs font-bold uppercase tracking-wider pt-2" style={{ color: 'var(--text-muted)' }}>Automation</p>
            <label className="flex items-center justify-between cursor-pointer">
              <div>
                <p className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>Auto-Reply (AI Conversation)</p>
                <p className="text-xs" style={{ color: 'var(--text-muted)' }}>Let AI respond to Messenger messages automatically</p>
              </div>
              <div className={`w-11 h-6 rounded-full relative transition-colors ${form.auto_reply_enabled ? 'bg-primary' : 'bg-gray-600'}`} onClick={() => setForm({ ...form, auto_reply_enabled: !form.auto_reply_enabled })}>
                <div className={`absolute top-0.5 left-0.5 w-5 h-5 rounded-full bg-white transition-transform ${form.auto_reply_enabled ? 'translate-x-5' : ''}`} />
              </div>
            </label>
            <label className="flex items-center justify-between cursor-pointer">
              <div>
                <p className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>Auto-Order Processing</p>
                <p className="text-xs" style={{ color: 'var(--text-muted)' }}>Automatically verify payments and book couriers</p>
              </div>
              <div className={`w-11 h-6 rounded-full relative transition-colors ${form.auto_order_enabled ? 'bg-primary' : 'bg-gray-600'}`} onClick={() => setForm({ ...form, auto_order_enabled: !form.auto_order_enabled })}>
                <div className={`absolute top-0.5 left-0.5 w-5 h-5 rounded-full bg-white transition-transform ${form.auto_order_enabled ? 'translate-x-5' : ''}`} />
              </div>
            </label>
          </div>

          {/* Save */}
          <button onClick={handleSave} disabled={saving} className="btn-primary flex items-center gap-2">
            <Save className="w-4 h-4" />
            {saving ? 'Saving...' : isNew ? 'Create Shop' : 'Save Changes'}
          </button>
        </motion.div>

		{/* ── AI Rules Section ── */}
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }} className="space-y-4">
          <div className="glass-card p-6 border border-primary/20 bg-gradient-to-r from-primary/10 to-transparent">
            <div className="flex items-start gap-4">
              <div className="p-3 bg-primary/20 rounded-xl shrink-0">
                <Brain className="w-8 h-8 text-primary" />
              </div>
              <div className="flex-1">
                <h2 className="text-xl font-heading font-bold dark:text-white text-gray-900 mb-2">
                  AI Rules Engine
                </h2>
                <p className="text-sm text-dark-300 mb-4">
                  Train your Xeni bot visually. Set custom rules for tone, delivery, return policy, and behavior alongside our platform-wide master constraints.
                </p>
                <Link href="/dashboard/rules" className="btn-primary inline-flex items-center gap-2">
                  <Store className="w-4 h-4" /> Manage AI Rules
                </Link>
              </div>
            </div>
          </div>
        </motion.div>

        {/* ── Payment API Integration ── */}
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.15 }} className="glass-card p-6">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-lg font-heading font-bold" style={{ color: 'var(--text-primary)' }}>Payment Verification Setup</h2>
              <p className="text-sm mt-1" style={{ color: 'var(--text-muted)' }}>Configure how order payments are verified by the AI.</p>
            </div>
            <select 
              className="input-field w-auto font-bold" 
              value={form.payment_verification_mode} 
              onChange={e => setForm({ ...form, payment_verification_mode: e.target.value })}
            >
              <option value="manual">👤 Manual Review</option>
              <option value="auto">🤖 Auto API Verification</option>
            </select>
          </div>

          {form.payment_verification_mode === 'manual' ? (
            <div className="p-4 rounded-xl bg-orange-500/10 border border-orange-500/20 text-orange-600 dark:text-orange-400 text-sm flex gap-3">
              <Info className="w-5 h-5 shrink-0" />
              <p>আপনি <strong>ম্যানুয়াল রিভিউ</strong> সিলেক্ট করেছেন। কাস্টমার পেমেন্ট ট্রানজেকশন আইডি বা স্ক্রিনশট দিলে সেটা আপনার ড্যাশবোর্ডে <strong>Manual Review</strong> ট্যাবে আসবে এবং আপনি হোয়াটসঅ্যাপে নোটিফিকেশন পাবেন। আপনাকে নিজে চেক করে Approve করতে হবে।</p>
            </div>
          ) : (
            <div className="space-y-6 animate-in slide-in-from-top-2">
              <div className="p-4 rounded-xl bg-primary/10 border border-primary/20 text-primary text-sm flex gap-3">
                <Brain className="w-5 h-5 shrink-0" />
                <p>আপনি <strong>অটো ভেরিফিকেশন</strong> সিলেক্ট করেছেন। দয়া করে নিচের API Key গুলো সঠিকভাবে দিন। কাস্টমার টাকা পাঠালে AI নিজে নিজে চেক করে অর্ডার কনফার্ম করে দিবে।</p>
              </div>

              {/* bKash Tokenized Checkout API */}
              <div className="p-4 rounded-xl shadow-sm border" style={{ borderColor: 'var(--border-color)', background: 'var(--bg-primary)' }}>
                <h3 className="font-semibold mb-3 flex items-center gap-2 text-pink-600 dark:text-pink-400">bKash Tokenized API Integration</h3>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs font-medium mb-1" style={{ color: 'var(--text-secondary)' }}>App Key</label>
                    <input className="input-field" placeholder="bKash App Key" value={form.bkash_app_key} onChange={e => setForm({ ...form, bkash_app_key: e.target.value })} />
                  </div>
                  <div>
                    <label className="block text-xs font-medium mb-1" style={{ color: 'var(--text-secondary)' }}>App Secret <span className="text-[10px] text-gray-400">(leave blank if unchanged)</span></label>
                    <input className="input-field" type="password" placeholder="bKash App Secret" value={form.bkash_app_secret} onChange={e => setForm({ ...form, bkash_app_secret: e.target.value })} />
                  </div>
                </div>
              </div>

              {/* Nagad API */}
              <div className="p-4 rounded-xl shadow-sm border" style={{ borderColor: 'var(--border-color)', background: 'var(--bg-primary)' }}>
                <h3 className="font-semibold mb-3 flex items-center gap-2 text-orange-600 dark:text-orange-400">Nagad API Integration</h3>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs font-medium mb-1" style={{ color: 'var(--text-secondary)' }}>Merchant ID</label>
                    <input className="input-field" placeholder="Nagad Merchant ID" value={form.nagad_merchant_id} onChange={e => setForm({ ...form, nagad_merchant_id: e.target.value })} />
                  </div>
                  <div>
                    <label className="block text-xs font-medium mb-1" style={{ color: 'var(--text-secondary)' }}>Merchant Key <span className="text-[10px] text-gray-400">(leave blank if unchanged)</span></label>
                    <input className="input-field" type="password" placeholder="Nagad Merchant Key" value={form.nagad_merchant_key} onChange={e => setForm({ ...form, nagad_merchant_key: e.target.value })} />
                  </div>
                </div>
              </div>

              {/* SSLCommerz Online Payment Gateway */}
              <div className="p-4 rounded-xl shadow-sm border" style={{ borderColor: 'var(--border-color)', background: 'var(--bg-primary)' }}>
                <div className="flex items-center justify-between mb-3">
                  <h3 className="font-semibold flex items-center gap-2 text-emerald-600 dark:text-emerald-400">
                    <Lock className="w-4 h-4" /> SSLCommerz Multi-Card Gateway
                  </h3>
                  <label className="flex items-center gap-2 cursor-pointer text-xs">
                    <span style={{ color: 'var(--text-muted)' }}>{form.sslcommerz_is_live ? '🔴 Live Mode' : '🟡 Sandbox Test'}</span>
                    <input
                      type="checkbox"
                      checked={form.sslcommerz_is_live}
                      onChange={e => setForm({ ...form, sslcommerz_is_live: e.target.checked })}
                      className="rounded"
                    />
                  </label>
                </div>
                <p className="text-xs mb-3" style={{ color: 'var(--text-muted)' }}>
                  E-Pic মার্কেটপ্লেসে কাস্টমার কার্ড বা নেট ব্যাংকিং দিয়ে পে করলে টাকা সরাসরি আপনার এই SSLCommerz অ্যাকাউন্টে জমা হবে।
                </p>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs font-medium mb-1" style={{ color: 'var(--text-secondary)' }}>Store ID</label>
                    <input className="input-field" placeholder="e.g. yourstore001" value={form.sslcommerz_store_id} onChange={e => setForm({ ...form, sslcommerz_store_id: e.target.value })} />
                  </div>
                  <div>
                    <label className="block text-xs font-medium mb-1" style={{ color: 'var(--text-secondary)' }}>Store Password <span className="text-[10px] text-gray-400">(leave blank if unchanged)</span></label>
                    <input className="input-field" type="password" placeholder="Store Password" value={form.sslcommerz_store_passwd} onChange={e => setForm({ ...form, sslcommerz_store_passwd: e.target.value })} />
                  </div>
                </div>
              </div>
            </div>
          )}
        </motion.div>

        {/* ── Merchant Courier API Keys ── */}
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.2 }} className="glass-card p-6">
          <h2 className="text-lg font-heading font-bold mb-2" style={{ color: 'var(--text-primary)' }}>Merchant Courier API Keys</h2>
          <p className="text-sm mb-6" style={{ color: 'var(--text-muted)' }}>Configure your custom connection keys if you want the AI to handle courier bookings under your own corporate account.</p>

          <div className="space-y-6">
            {/* Pathao API block */}
            <div className="p-4 rounded-xl shadow-sm border" style={{ borderColor: 'var(--border-color)', background: 'var(--bg-primary)' }}>
              <h3 className="font-semibold mb-3 flex items-center gap-2 text-primary">Pathao Integration</h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-medium mb-1" style={{ color: 'var(--text-secondary)' }}>Client ID</label>
                   <input className="input-field" placeholder="Client ID" value={form.integrations.pathao_client_id} onChange={e => setForm({ ...form, integrations: { ...form.integrations, pathao_client_id: e.target.value } })} />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1" style={{ color: 'var(--text-secondary)' }}>Client Secret</label>
                   <input className="input-field" type="password" placeholder="Client Secret" value={form.integrations.pathao_client_secret} onChange={e => setForm({ ...form, integrations: { ...form.integrations, pathao_client_secret: e.target.value } })} />
                </div>
              </div>
            </div>

            {/* Steadfast API block */}
            <div className="p-4 rounded-xl shadow-sm border" style={{ borderColor: 'var(--border-color)', background: 'var(--bg-primary)' }}>
              <h3 className="font-semibold mb-3 flex items-center gap-2 text-primary">Steadfast Integration</h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-medium mb-1" style={{ color: 'var(--text-secondary)' }}>API Key</label>
                   <input className="input-field" placeholder="API Key" value={form.integrations.steadfast_api_key} onChange={e => setForm({ ...form, integrations: { ...form.integrations, steadfast_api_key: e.target.value } })} />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1" style={{ color: 'var(--text-secondary)' }}>Secret Key</label>
                   <input className="input-field" type="password" placeholder="Secret Key" value={form.integrations.steadfast_secret_key} onChange={e => setForm({ ...form, integrations: { ...form.integrations, steadfast_secret_key: e.target.value } })} />
                </div>
              </div>
            </div>
          </div>

          <button onClick={handleSave} disabled={saving} className="btn-primary mt-6 flex items-center gap-2">
            <Save className="w-4 h-4" />
            {saving ? 'Saving...' : 'Save Integrations'}
          </button>
        </motion.div>
      </div>
    </div>
  );
}
