import {useEffect, useState} from 'react';
import {Alert, Box, Button, Card, CardContent, Divider, TextField, Typography} from '@mui/material';
import {HStack, VStack} from "@dsherwin/mui-kit";
import {formatAPIError} from "../../lib/errors";
import {SectionHelp} from '../../main/SectionHelp';
import {HELP_SECTIONS} from '../../main/appShellManifest.ts';
import {fetchBrandingSettings, saveBrandingSettings} from './api.ts';

export const BrandingPage = () => {
    const [brandLogoUrl, setBrandLogoUrl] = useState('');
    const [brandName, setBrandName] = useState('');
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState(false);

    const load = async () => {
        setLoading(true);
        try {
            const data = await fetchBrandingSettings();
            setBrandLogoUrl(data.brandLogoUrl || '');
            setBrandName(data.brandName || '');
        } catch (e) {
            console.error(e);
            setError('Failed to load branding settings.');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        load();
    }, []);

    const onSave = async () => {
        setSaving(true);
        setError(null);
        setSuccess(false);
        try {
            await saveBrandingSettings({brandLogoUrl, brandName});
            setSuccess(true);
        } catch (e) {
            console.error(e);
            setError(formatAPIError(e, 'Failed to save branding settings.'));
        } finally {
            setSaving(false);
        }
    };

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={3} sx={{maxWidth: 800, width: '100%', mx: 'auto'}}>
                <Box sx={{display: 'flex', alignItems: 'center'}}>
                    <Typography variant="h5">Branding</Typography>
                    <SectionHelp section={HELP_SECTIONS.activity} />
                </Box>
                <Divider/>

                {error && <Alert severity="error" onClose={() => setError(null)}>{error}</Alert>}
                {success && <Alert severity="success" onClose={() => setSuccess(false)}>Branding settings saved successfully.</Alert>}

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardContent>
                        <VStack spacing={3}>
                            <Box>
                                <Typography variant="subtitle2" gutterBottom sx={{fontWeight: 'bold'}}>Custom Logo URL</Typography>
                                <Typography
                                    variant="body2"
                                    sx={{
                                        color: "text.secondary",
                                        mb: 2
                                    }}>
                                    Provide a URL to your custom logo. This will replace the Chronix logo in the top navigation bar.
                                    Recommended format: SVG or transparent PNG, max height 40px.
                                </Typography>
                                <TextField
                                    fullWidth
                                    placeholder="https://example.com/logo.svg"
                                    value={brandLogoUrl}
                                    onChange={(e) => setBrandLogoUrl(e.target.value)}
                                />
                            </Box>

                            <Box>
                                <Typography variant="subtitle2" gutterBottom sx={{fontWeight: 'bold'}}>Custom Brand Name</Typography>
                                <Typography
                                    variant="body2"
                                    sx={{
                                        color: "text.secondary",
                                        mb: 2
                                    }}>
                                    This name will be used in browser titles and other UI elements if enabled.
                                </Typography>
                                <TextField
                                    fullWidth
                                    placeholder="My Company"
                                    value={brandName}
                                    onChange={(e) => setBrandName(e.target.value)}
                                />
                            </Box>

                            <Box sx={{pt: 1}}>
                                <Button
                                    variant="contained"
                                    onClick={onSave}
                                    disabled={saving || loading}
                                >
                                    {saving ? 'Saving...' : 'Save Branding Settings'}
                                </Button>
                            </Box>
                        </VStack>
                    </CardContent>
                </Card>

                {(brandLogoUrl || brandName) && (
                    <Box>
                        <Typography variant="subtitle2" gutterBottom sx={{fontWeight: 'bold'}}>Preview</Typography>
                        <Card variant="outlined" sx={{borderRadius: 3, bgcolor: (theme) => theme.palette.mode === 'dark' ? '#1e1e1e' : '#42A5F5', p: 2}}>
                            <HStack alignItems="center" spacing={2}>
                                {brandLogoUrl ? (
                                    <img src={brandLogoUrl} alt="Logo Preview" style={{height: 40, maxWidth: 150, objectFit: 'contain'}} />
                                ) : (
                                    <Typography variant="h6" color="white">{brandName || 'Chronix'}</Typography>
                                )}
                            </HStack>
                        </Card>
                    </Box>
                )}
            </VStack>
        </Box>
    );
};
