package events

const (
	UserCreatedEventName             = "user.created"
	UserProfileUpdatedEventName      = "user.profileUpdated"
	UserContactUpdatedEventName      = "user.contactUpdated"
	UserPasswordChangedEventName     = "user.passwordChanged"
	UserStatusChangedEventName       = "user.statusChanged"
	UserEmailVerifiedEventName       = "user.emailVerified"
	UserPhoneVerifiedEventName       = "user.phoneVerified"
	UserAuthProviderLinkedEventName  = "user.authProviderLinked"
	UserAuthProviderUsedEventName    = "user.authProviderUsed"
	UserOnboardingCompletedEventName = "user.onboardingCompleted"
	UserLoggedInEventName            = "user.loggedIn"

	MedicationCreatedEventName    = "medication.created"
	MedicationUpdatedEventName    = "medication.updated"
	MedicationTimeAddedEventName  = "medication.timeAdded"
	MedicationDoseLoggedEventName = "medication.doseLogged"
	MedicationCompletedEventName  = "medication.completed"
	MedicationAdherenceEventName  = "medication.adherence_logged"

	VisitCreatedEventName            = "visit.created"
	VisitUpdatedEventName            = "visit.updated"
	VisitMedicationAttachedEventName = "visit.medicationAttached"

	AIConversationCreatedEventName  = "aiConversation.created"
	AIMessageAddedEventName         = "aiConversation.messageAdded"
	AIConversationArchivedEventName = "aiConversation.archived"

	DrugScanCreatedEventName  = "drugScan.created"
	DrugScanVerifiedEventName = "drugScan.verified"
	DrugScanRejectedEventName = "drugScan.rejected"

	NotificationCreatedEventName = "notification.created"
	NotificationReadEventName    = "notification.read"
	NotificationSentEventName    = "notification.sent"
	NotificationFailedEventName  = "notification.failed"

	ScheduledJobCreatedEventName   = "scheduledNotificationJob.created"
	ScheduledJobPausedEventName    = "scheduledNotificationJob.paused"
	ScheduledJobCancelledEventName = "scheduledNotificationJob.cancelled"

	SupportTicketCreatedEventName      = "supportTicket.created"
	SupportTicketMessageAddedEventName = "supportTicket.messageAdded"
	SupportTicketClosedEventName       = "supportTicket.closed"

	FeedbackCreatedEventName    = "feedback.created"
	FeedbackReplyAddedEventName = "feedback.replyAdded"
	FeedbackResolvedEventName   = "feedback.resolved"

	RegisteredMedicineCreatedEventName       = "registeredMedicine.created"
	RegisteredMedicineStatusChangedEventName = "registeredMedicine.statusChanged"
)
