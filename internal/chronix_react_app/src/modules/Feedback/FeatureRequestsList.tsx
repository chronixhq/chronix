import React from 'react';
import {FeedbackList} from './FeedbackList';

export const FeatureRequestsList: React.FC = () => {
    return (
        <FeedbackList 
            title="Feature Requests" 
            kind="feature" 
        />
    );
};
