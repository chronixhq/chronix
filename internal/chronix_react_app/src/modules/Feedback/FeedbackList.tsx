import React, {useEffect, useState} from 'react';
import {Accordion, AccordionDetails, AccordionSummary, Alert, Box, Button, Chip, CircularProgress, Divider, FormControl, IconButton, InputLabel, Link, MenuItem, Select, TextField, Typography} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import AttachmentIcon from '@mui/icons-material/Attachment';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import SaveIcon from '@mui/icons-material/Save';
import CancelIcon from '@mui/icons-material/Cancel';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import {HStack, VStack} from "@dsherwin/mui-kit";
import {apiDelete, apiGet, apiPatch, apiPost} from '@dsherwin/react-api-interface';
import {baseURL} from '../../lib/api_config';
import type {FeedbackItem} from '../../data/types.ts';
import {formatDateTime} from '../../lib/utilities.tsx';
import {formatAPIError} from '../../lib/errors.ts';

const FeedbackItemComponent: React.FC<{
    item: FeedbackItem;
    kind: 'bug' | 'feature';
    onUpdate: () => void;
    onDelete: () => void;
}> = ({item, kind, onUpdate, onDelete}) => {
    const [isEditing, setIsEditing] = useState(false);
    const [summary, setSummary] = useState(item.summary);
    const [description, setDescription] = useState(item.description);
    const [status, setStatus] = useState(item.status);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const getStatusColor = (s: string) => {
        switch (s) {
            case 'open': return 'error';
            case 'in-progress': return 'warning';
            case 'closed': return 'success';
            default: return 'default';
        }
    };

    const handleSave = async () => {
        setSaving(true);
        setError(null);
        try {
            const endpoint = `/feedback/${kind === 'bug' ? 'bug-report' : 'feature-request'}/${item.id}`;
            await apiPatch(endpoint, {summary, description, status});
            setIsEditing(false);
            onUpdate();
        } catch (e: any) {
            console.error(e);
            setError(formatAPIError(e, 'Failed to update feedback.'));
        } finally {
            setSaving(false);
        }
    };

    const handleDelete = async () => {
        if (!window.confirm(`Are you sure you want to delete this ${kind === 'bug' ? 'bug report' : 'feature request'}?`)) {
            return;
        }
        try {
            const endpoint = `/feedback/${kind === 'bug' ? 'bug-report' : 'feature-request'}/${item.id}`;
            await apiDelete(endpoint);
            onDelete();
        } catch (e: any) {
            console.error(e);
            setError(formatAPIError(e, 'Failed to delete feedback.'));
        }
    };

    const handleDeleteAttachment = async (attId: number) => {
        if (!window.confirm('Are you sure you want to delete this attachment?')) {
            return;
        }
        try {
            await apiDelete(`/feedback/attachments/${attId}`);
            onUpdate();
        } catch (e: any) {
            console.error(e);
            setError(formatAPIError(e, 'Failed to delete attachment.'));
        }
    };

    const handleAddAttachments = async (e: React.ChangeEvent<HTMLInputElement>) => {
        if (!e.target.files || e.target.files.length === 0) return;
        
        const formData = new FormData();
        Array.from(e.target.files).forEach(file => {
            formData.append('attachments', file);
        });

        try {
            const endpoint = `/feedback/${kind === 'bug' ? 'bug-report' : 'feature-request'}/${item.id}/attachments`;
            await apiPost(endpoint, formData);
            onUpdate();
        } catch (e: any) {
            console.error(e);
            setError(formatAPIError(e, 'Failed to upload attachments.'));
        } finally {
            e.target.value = '';
        }
    };

    return (
        <Accordion variant="outlined" sx={{borderRadius: 2, overflow: 'hidden'}}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                <HStack justifyContent="space-between" sx={{width: '100%', mr: 2}}>
                    <VStack gap={0.5}>
                        <Typography variant="subtitle1" sx={{
                            fontWeight: "bold"
                        }}>{item.summary}</Typography>
                        <Typography variant="caption" sx={{
                            color: "text.secondary"
                        }}>
                            Submitted on {formatDateTime(item.createdAt)}
                        </Typography>
                    </VStack>
                    <Chip 
                        label={item.status.toUpperCase()} 
                        size="small" 
                        color={getStatusColor(item.status) as any}
                    />
                </HStack>
            </AccordionSummary>
            <AccordionDetails>
                <VStack spacing={3} sx={{width: '100%'}}>
                    {error && <Alert severity="error" onClose={() => setError(null)}>{error}</Alert>}
                    
                    {isEditing ? (
                        <VStack spacing={2}>
                            <TextField
                                fullWidth
                                label="Summary"
                                value={summary}
                                onChange={(e) => setSummary(e.target.value)}
                            />
                            <TextField
                                fullWidth
                                label="Description"
                                multiline
                                rows={4}
                                value={description}
                                onChange={(e) => setDescription(e.target.value)}
                            />
                            <FormControl fullWidth>
                                <InputLabel>Status</InputLabel>
                                <Select
                                    value={status}
                                    label="Status"
                                    onChange={(e) => setStatus(e.target.value as any)}
                                >
                                    <MenuItem value="open">Open</MenuItem>
                                    <MenuItem value="in-progress">In Progress</MenuItem>
                                    <MenuItem value="closed">Closed</MenuItem>
                                </Select>
                            </FormControl>
                            <HStack spacing={1}>
                                <Button 
                                    variant="contained" 
                                    startIcon={<SaveIcon />} 
                                    onClick={handleSave}
                                    disabled={saving}
                                >
                                    {saving ? 'Saving...' : 'Save'}
                                </Button>
                                <Button 
                                    variant="outlined" 
                                    startIcon={<CancelIcon />} 
                                    onClick={() => {
                                        setIsEditing(false);
                                        setSummary(item.summary);
                                        setDescription(item.description);
                                        setStatus(item.status);
                                    }}
                                    disabled={saving}
                                >
                                    Cancel
                                </Button>
                            </HStack>
                        </VStack>
                    ) : (
                        <>
                            <Typography variant="body2" sx={{whiteSpace: 'pre-wrap'}}>
                                {item.description}
                            </Typography>
                            <HStack spacing={1}>
                                <Button 
                                    size="small" 
                                    startIcon={<EditIcon />} 
                                    onClick={() => setIsEditing(true)}
                                >
                                    Edit
                                </Button>
                                <Button 
                                    size="small" 
                                    color="error" 
                                    startIcon={<DeleteIcon />} 
                                    onClick={handleDelete}
                                >
                                    Delete
                                </Button>
                            </HStack>
                        </>
                    )}
                    
                    <Divider />
                    
                    <Box>
                        <HStack justifyContent="space-between" alignItems="center" sx={{mb: 1}}>
                            <Typography variant="subtitle2" sx={{display: 'flex', alignItems: 'center'}}>
                                <AttachmentIcon fontSize="small" sx={{mr: 0.5}} />
                                Attachments
                            </Typography>
                            <Button
                                size="small"
                                component="label"
                                startIcon={<CloudUploadIcon />}
                            >
                                Add
                                <input
                                    type="file"
                                    hidden
                                    multiple
                                    onChange={handleAddAttachments}
                                />
                            </Button>
                        </HStack>
                        {item.feedbackAttachments && item.feedbackAttachments.length > 0 ? (
                            <HStack spacing={2} flexWrap="wrap">
                                {item.feedbackAttachments.map((att) => (
                                    <HStack 
                                        key={att.id}
                                        sx={{
                                            p: 0.5, 
                                            pl: 1,
                                            border: '1px solid',
                                            borderColor: 'divider',
                                            borderRadius: 1,
                                            bgcolor: 'background.paper',
                                            '&:hover': { bgcolor: 'action.hover' }
                                        }}
                                    >
                                        <Link 
                                            href={`${baseURL}/feedback/attachments/${att.id}`}
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            variant="caption"
                                            sx={{ textDecoration: 'none', display: 'flex', alignItems: 'center' }}
                                        >
                                            {att.fileName} ({(att.fileSize / 1024).toFixed(1)} KB)
                                        </Link>
                                        <IconButton size="small" onClick={() => handleDeleteAttachment(att.id)}>
                                            <DeleteIcon fontSize="inherit" />
                                        </IconButton>
                                    </HStack>
                                ))}
                            </HStack>
                        ) : (
                            <Typography variant="caption" sx={{
                                color: "text.secondary"
                            }}>No attachments</Typography>
                        )}
                    </Box>
                </VStack>
            </AccordionDetails>
        </Accordion>
    );
};

interface FeedbackListProps {
    title: string;
    kind: 'bug' | 'feature';
}

export const FeedbackList: React.FC<FeedbackListProps> = ({title, kind}) => {
    const [items, setItems] = useState<FeedbackItem[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const load = async () => {
        setLoading(true);
        setError(null);
        const endpoint = kind === 'bug' ? '/feedback/bug-reports' : '/feedback/feature-requests';
        try {
            const data = await apiGet(endpoint) as FeedbackItem[];
            setItems(data || []);
        } catch (e: any) {
            console.error(e);
            setError('Failed to load feedback items.');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        void load();
    }, [kind]); // eslint-disable-line react-hooks/exhaustive-deps

    if (loading) {
        return (
            <Box sx={{display: 'flex', justifyContent: 'center', py: 8}}>
                <CircularProgress />
            </Box>
        );
    }

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={3} sx={{maxWidth: 1000, width: '100%', mx: 'auto'}}>
                <Typography variant="h5">{title}</Typography>
                <Divider/>

                {error && <Alert severity="error">{error}</Alert>}

                {items.length === 0 ? (
                    <Typography sx={{
                        color: "text.secondary"
                    }}>No {kind === 'bug' ? 'bug reports' : 'feature requests'} found.</Typography>
                ) : (
                    items.map((item) => (
                        <FeedbackItemComponent 
                            key={item.id} 
                            item={item} 
                            kind={kind} 
                            onUpdate={load} 
                            onDelete={load} 
                        />
                    ))
                )}
            </VStack>
        </Box>
    );
};
