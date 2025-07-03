import React, { useState } from 'react';
import { TagStandard } from '../types';
import StandardsList from '../components/standards/StandardsList';
import StandardEditor from '../components/standards/StandardEditor';

const StandardsPage: React.FC = () => {
  const [isEditorOpen, setIsEditorOpen] = useState(false);
  const [selectedStandard, setSelectedStandard] = useState<TagStandard | null>(null);

  const handleCreateNew = () => {
    setSelectedStandard(null);
    setIsEditorOpen(true);
  };

  const handleEdit = (standard: TagStandard) => {
    setSelectedStandard(standard);
    setIsEditorOpen(true);
  };

  const handleView = (standard: TagStandard) => {
    // TODO: Implement view mode
    console.log('View standard:', standard);
  };

  const handleCloseEditor = () => {
    setIsEditorOpen(false);
    setSelectedStandard(null);
  };

  return (
    <>
      <StandardsList 
        onCreateNew={handleCreateNew}
        onEdit={handleEdit}
        onView={handleView}
      />
      <StandardEditor
        isOpen={isEditorOpen}
        onClose={handleCloseEditor}
        standard={selectedStandard}
      />
    </>
  );
};

export default StandardsPage;