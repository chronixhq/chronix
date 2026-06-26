import React from 'react';
import {FeedbackForm} from './FeedbackForm';
import {useNavigate} from 'react-router';

export const FeatureRequestPage: React.FC = () => {
    const navigate = useNavigate();
    return (
        <FeedbackForm 
            title="Request a Feature" 
            kind="feature" 
            onSubmitSuccess={() => navigate('/')} 
        />
    );
};
