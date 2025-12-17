-- +goose Up
-- =============================================================================
-- Translate Activity Titles and Bodies from English to Norwegian
-- =============================================================================
-- This migration updates existing activity records to use Norwegian text
-- to match the updated service layer translations.
-- =============================================================================

-- Customer activities
UPDATE activities SET title = 'Kunde opprettet' WHERE title = 'Customer created';
UPDATE activities SET title = 'Kunde oppdatert' WHERE title = 'Customer updated';
UPDATE activities SET title = 'Kunde slettet' WHERE title = 'Customer deleted';
UPDATE activities SET title = 'Status oppdatert' WHERE title = 'Status updated';
UPDATE activities SET title = 'Nivå oppdatert' WHERE title = 'Level updated';
UPDATE activities SET title = 'Bransje oppdatert' WHERE title = 'Industry updated';
UPDATE activities SET title = 'Notater oppdatert' WHERE title = 'Notes updated';
UPDATE activities SET title = 'Selskap oppdatert' WHERE title = 'Company updated';
UPDATE activities SET title = 'Kundeklasse oppdatert' WHERE title = 'Customer class updated';
UPDATE activities SET title = 'Kredittgrense oppdatert' WHERE title = 'Credit limit updated';
UPDATE activities SET title = 'Intern-flagg oppdatert' WHERE title = 'Internal flag updated';
UPDATE activities SET title = 'Adresse oppdatert' WHERE title = 'Address updated';
UPDATE activities SET title = 'Postnummer oppdatert' WHERE title = 'Postal code updated';
UPDATE activities SET title = 'By oppdatert' WHERE title = 'City updated';
UPDATE activities SET title = 'Kontaktinfo oppdatert' WHERE title = 'Contact info updated';

-- Contact activities
UPDATE activities SET title = 'Kontakt opprettet' WHERE title = 'Contact created';
UPDATE activities SET title = 'Kontakt oppdatert' WHERE title = 'Contact updated';
UPDATE activities SET title = 'Kontakt slettet' WHERE title = 'Contact deleted';
UPDATE activities SET title = 'Relasjon lagt til' WHERE title = 'Relationship added';
UPDATE activities SET title = 'Relasjon fjernet' WHERE title = 'Relationship removed';

-- Offer activities
UPDATE activities SET title = 'Tilbud opprettet' WHERE title = 'Offer created';
UPDATE activities SET title = 'Tilbud oppdatert' WHERE title = 'Offer updated';
UPDATE activities SET title = 'Tilbud slettet' WHERE title = 'Offer deleted';
UPDATE activities SET title = 'Tilbud sendt' WHERE title = 'Offer sent';
UPDATE activities SET title = 'Ordre akseptert' WHERE title = 'Order accepted';
UPDATE activities SET title = 'Tilbudshelse oppdatert' WHERE title = 'Offer health updated';
UPDATE activities SET title = 'Tilbudsforbruk oppdatert' WHERE title = 'Offer consumption updated';
UPDATE activities SET title = 'Tilbud fakturering oppdatert' WHERE title = 'Offer billing updated';
UPDATE activities SET title = 'Tilbud fullført' WHERE title = 'Offer completed';
UPDATE activities SET title = 'Tilbud avslått' WHERE title = 'Offer lost';
UPDATE activities SET title = 'Tilbud utløpt' WHERE title = 'Offer expired';
UPDATE activities SET title = 'Tilbud tilbakestilt til sendt' WHERE title = 'Offer reset to sent';
UPDATE activities SET title = 'Tilbud klonet' WHERE title = 'Offer cloned';
UPDATE activities SET title = 'Tilbud opprettet fra klone' WHERE title = 'Offer created from clone';
UPDATE activities SET title = 'Tilbudstotaler omberegnet' WHERE title = 'Offer totals recalculated';
UPDATE activities SET title = 'Tilbudsfase avansert' WHERE title = 'Offer phase advanced';
UPDATE activities SET title = 'Prosjekt auto-opprettet' WHERE title = 'Project auto-created';
UPDATE activities SET title = 'Tilbudssannsynlighet oppdatert' WHERE title = 'Offer probability updated';
UPDATE activities SET title = 'Tilbudstittel oppdatert' WHERE title = 'Offer title updated';
UPDATE activities SET title = 'Tilbudsansvarlig oppdatert' WHERE title = 'Offer manager updated';
UPDATE activities SET title = 'Tilbudskunde oppdatert' WHERE title = 'Offer customer updated';
UPDATE activities SET title = 'Tilbudsverdi oppdatert' WHERE title = 'Offer value updated';
UPDATE activities SET title = 'Tilbudskostnad oppdatert' WHERE title = 'Offer cost updated';
UPDATE activities SET title = 'Tilbud forfallsdato oppdatert' WHERE title = 'Offer due date updated';
UPDATE activities SET title = 'Tilbud utløpsdato oppdatert' WHERE title = 'Offer expiry date updated';
UPDATE activities SET title = 'Tilbudsbeskrivelse oppdatert' WHERE title = 'Offer description updated';
UPDATE activities SET title = 'Tilbud koblet til prosjekt' WHERE title = 'Offer linked to project';
UPDATE activities SET title = 'Tilbud frakoblet fra prosjekt' WHERE title = 'Offer unlinked from project';
UPDATE activities SET title = 'Kundens prosjektstatus oppdatert' WHERE title = 'Customer project status updated';
UPDATE activities SET title = 'Tilbudsnummer oppdatert' WHERE title = 'Offer number updated';
UPDATE activities SET title = 'Ekstern referanse oppdatert' WHERE title = 'External reference updated';

-- Project activities
UPDATE activities SET title = 'Prosjekt opprettet' WHERE title = 'Project created';
UPDATE activities SET title = 'Prosjekt oppdatert' WHERE title = 'Project updated';
UPDATE activities SET title = 'Prosjekt slettet' WHERE title = 'Project deleted';
UPDATE activities SET title = 'Prosjektfase oppdatert' WHERE title = 'Project phase updated';
UPDATE activities SET title = 'Prosjekt gjenåpnet' WHERE title = 'Project reopened';
UPDATE activities SET title = 'Prosjektnavn oppdatert' WHERE title = 'Project name updated';
UPDATE activities SET title = 'Prosjektbeskrivelse oppdatert' WHERE title = 'Project description updated';
UPDATE activities SET title = 'Prosjektdatoer oppdatert' WHERE title = 'Project dates updated';
UPDATE activities SET title = 'Prosjektnummer oppdatert' WHERE title = 'Project number updated';

-- Deal activities
UPDATE activities SET title = 'Salgsmulighet opprettet' WHERE title = 'Deal created';
UPDATE activities SET title = 'Salgsmulighet oppdatert' WHERE title = 'Deal updated';
UPDATE activities SET title = 'Salgsmulighet fase endret' WHERE title = 'Deal stage changed';
UPDATE activities SET title = 'Salgsmulighet vunnet!' WHERE title = 'Deal won!';
UPDATE activities SET title = 'Prosjekt opprettet fra salgsmulighet' WHERE title = 'Project created from deal';
UPDATE activities SET title = 'Salgsmulighet tapt' WHERE title = 'Deal lost';
UPDATE activities SET title = 'Salgsmulighet gjenåpnet' WHERE title = 'Deal reopened';
UPDATE activities SET title = 'Tilbud opprettet fra salgsmulighet' WHERE title = 'Offer created from deal';

-- File activities
UPDATE activities SET title = 'Fil lastet opp' WHERE title = 'File uploaded';

-- Role activities (pattern: "Role assigned", "Role removed")
UPDATE activities SET title = 'Rolle tildelt' WHERE title = 'Role assigned';
UPDATE activities SET title = 'Rolle fjernet' WHERE title = 'Role removed';

-- Permission activities (pattern: "Permission granted", etc.)
UPDATE activities SET title = 'Tillatelse gitt' WHERE title = 'Permission granted';
UPDATE activities SET title = 'Tillatelse nektet' WHERE title = 'Permission denied';
UPDATE activities SET title = 'Tillatelse fjernet (overstyring)' WHERE title = 'Permission override removed';

-- +goose Down
-- Revert Norwegian titles back to English

-- Customer activities
UPDATE activities SET title = 'Customer created' WHERE title = 'Kunde opprettet';
UPDATE activities SET title = 'Customer updated' WHERE title = 'Kunde oppdatert';
UPDATE activities SET title = 'Customer deleted' WHERE title = 'Kunde slettet';
UPDATE activities SET title = 'Status updated' WHERE title = 'Status oppdatert';
UPDATE activities SET title = 'Level updated' WHERE title = 'Nivå oppdatert';
UPDATE activities SET title = 'Industry updated' WHERE title = 'Bransje oppdatert';
UPDATE activities SET title = 'Notes updated' WHERE title = 'Notater oppdatert';
UPDATE activities SET title = 'Company updated' WHERE title = 'Selskap oppdatert';
UPDATE activities SET title = 'Customer class updated' WHERE title = 'Kundeklasse oppdatert';
UPDATE activities SET title = 'Credit limit updated' WHERE title = 'Kredittgrense oppdatert';
UPDATE activities SET title = 'Internal flag updated' WHERE title = 'Intern-flagg oppdatert';
UPDATE activities SET title = 'Address updated' WHERE title = 'Adresse oppdatert';
UPDATE activities SET title = 'Postal code updated' WHERE title = 'Postnummer oppdatert';
UPDATE activities SET title = 'City updated' WHERE title = 'By oppdatert';
UPDATE activities SET title = 'Contact info updated' WHERE title = 'Kontaktinfo oppdatert';

-- Contact activities
UPDATE activities SET title = 'Contact created' WHERE title = 'Kontakt opprettet';
UPDATE activities SET title = 'Contact updated' WHERE title = 'Kontakt oppdatert';
UPDATE activities SET title = 'Contact deleted' WHERE title = 'Kontakt slettet';
UPDATE activities SET title = 'Relationship added' WHERE title = 'Relasjon lagt til';
UPDATE activities SET title = 'Relationship removed' WHERE title = 'Relasjon fjernet';

-- Offer activities
UPDATE activities SET title = 'Offer created' WHERE title = 'Tilbud opprettet';
UPDATE activities SET title = 'Offer updated' WHERE title = 'Tilbud oppdatert';
UPDATE activities SET title = 'Offer deleted' WHERE title = 'Tilbud slettet';
UPDATE activities SET title = 'Offer sent' WHERE title = 'Tilbud sendt';
UPDATE activities SET title = 'Order accepted' WHERE title = 'Ordre akseptert';
UPDATE activities SET title = 'Offer health updated' WHERE title = 'Tilbudshelse oppdatert';
UPDATE activities SET title = 'Offer consumption updated' WHERE title = 'Tilbudsforbruk oppdatert';
UPDATE activities SET title = 'Offer billing updated' WHERE title = 'Tilbud fakturering oppdatert';
UPDATE activities SET title = 'Offer completed' WHERE title = 'Tilbud fullført';
UPDATE activities SET title = 'Offer lost' WHERE title = 'Tilbud avslått';
UPDATE activities SET title = 'Offer expired' WHERE title = 'Tilbud utløpt';
UPDATE activities SET title = 'Offer reset to sent' WHERE title = 'Tilbud tilbakestilt til sendt';
UPDATE activities SET title = 'Offer cloned' WHERE title = 'Tilbud klonet';
UPDATE activities SET title = 'Offer created from clone' WHERE title = 'Tilbud opprettet fra klone';
UPDATE activities SET title = 'Offer totals recalculated' WHERE title = 'Tilbudstotaler omberegnet';
UPDATE activities SET title = 'Offer phase advanced' WHERE title = 'Tilbudsfase avansert';
UPDATE activities SET title = 'Project auto-created' WHERE title = 'Prosjekt auto-opprettet';
UPDATE activities SET title = 'Offer probability updated' WHERE title = 'Tilbudssannsynlighet oppdatert';
UPDATE activities SET title = 'Offer title updated' WHERE title = 'Tilbudstittel oppdatert';
UPDATE activities SET title = 'Offer manager updated' WHERE title = 'Tilbudsansvarlig oppdatert';
UPDATE activities SET title = 'Offer customer updated' WHERE title = 'Tilbudskunde oppdatert';
UPDATE activities SET title = 'Offer value updated' WHERE title = 'Tilbudsverdi oppdatert';
UPDATE activities SET title = 'Offer cost updated' WHERE title = 'Tilbudskostnad oppdatert';
UPDATE activities SET title = 'Offer due date updated' WHERE title = 'Tilbud forfallsdato oppdatert';
UPDATE activities SET title = 'Offer expiry date updated' WHERE title = 'Tilbud utløpsdato oppdatert';
UPDATE activities SET title = 'Offer description updated' WHERE title = 'Tilbudsbeskrivelse oppdatert';
UPDATE activities SET title = 'Offer linked to project' WHERE title = 'Tilbud koblet til prosjekt';
UPDATE activities SET title = 'Offer unlinked from project' WHERE title = 'Tilbud frakoblet fra prosjekt';
UPDATE activities SET title = 'Customer project status updated' WHERE title = 'Kundens prosjektstatus oppdatert';
UPDATE activities SET title = 'Offer number updated' WHERE title = 'Tilbudsnummer oppdatert';
UPDATE activities SET title = 'External reference updated' WHERE title = 'Ekstern referanse oppdatert';

-- Project activities
UPDATE activities SET title = 'Project created' WHERE title = 'Prosjekt opprettet';
UPDATE activities SET title = 'Project updated' WHERE title = 'Prosjekt oppdatert';
UPDATE activities SET title = 'Project deleted' WHERE title = 'Prosjekt slettet';
UPDATE activities SET title = 'Project phase updated' WHERE title = 'Prosjektfase oppdatert';
UPDATE activities SET title = 'Project reopened' WHERE title = 'Prosjekt gjenåpnet';
UPDATE activities SET title = 'Project name updated' WHERE title = 'Prosjektnavn oppdatert';
UPDATE activities SET title = 'Project description updated' WHERE title = 'Prosjektbeskrivelse oppdatert';
UPDATE activities SET title = 'Project dates updated' WHERE title = 'Prosjektdatoer oppdatert';
UPDATE activities SET title = 'Project number updated' WHERE title = 'Prosjektnummer oppdatert';

-- Deal activities
UPDATE activities SET title = 'Deal created' WHERE title = 'Salgsmulighet opprettet';
UPDATE activities SET title = 'Deal updated' WHERE title = 'Salgsmulighet oppdatert';
UPDATE activities SET title = 'Deal stage changed' WHERE title = 'Salgsmulighet fase endret';
UPDATE activities SET title = 'Deal won!' WHERE title = 'Salgsmulighet vunnet!';
UPDATE activities SET title = 'Project created from deal' WHERE title = 'Prosjekt opprettet fra salgsmulighet';
UPDATE activities SET title = 'Deal lost' WHERE title = 'Salgsmulighet tapt';
UPDATE activities SET title = 'Deal reopened' WHERE title = 'Salgsmulighet gjenåpnet';
UPDATE activities SET title = 'Offer created from deal' WHERE title = 'Tilbud opprettet fra salgsmulighet';

-- File activities
UPDATE activities SET title = 'File uploaded' WHERE title = 'Fil lastet opp';

-- Role activities
UPDATE activities SET title = 'Role assigned' WHERE title = 'Rolle tildelt';
UPDATE activities SET title = 'Role removed' WHERE title = 'Rolle fjernet';

-- Permission activities
UPDATE activities SET title = 'Permission granted' WHERE title = 'Tillatelse gitt';
UPDATE activities SET title = 'Permission denied' WHERE title = 'Tillatelse nektet';
UPDATE activities SET title = 'Permission override removed' WHERE title = 'Tillatelse fjernet (overstyring)';
