import greet from "./tools/greet";
import listEmails from "./tools/list_emails";
import getEmail from "./tools/get_email";
import sendEmail from "./tools/send_email";
import draftEmail from "./tools/draft_email";
import searchEmails from "./tools/search_emails";
import listLabels from "./tools/list_labels";
import modifyEmail from "./tools/modify_email";
import deleteEmail from "./tools/delete_email";
import batchModifyEmails from "./tools/batch_modify_emails";
import batchDeleteEmails from "./tools/batch_delete_emails";
import createLabel from "./tools/create_label";
import updateLabel from "./tools/update_label";
import deleteLabel from "./tools/delete_label";
import getOrCreateLabel from "./tools/get_or_create_label";
import downloadAttachment from "./tools/download_attachment";

export default {
  // Basic tools
  greet,
  
  // Email management
  list_emails: listEmails,
  get_email: getEmail,
  send_email: sendEmail,
  draft_email: draftEmail,
  search_emails: searchEmails,
  
  // Email modification
  modify_email: modifyEmail,
  delete_email: deleteEmail,
  batch_modify_emails: batchModifyEmails,
  batch_delete_emails: batchDeleteEmails,
  
  // Label management
  list_labels: listLabels,
  create_label: createLabel,
  update_label: updateLabel,
  delete_label: deleteLabel,
  get_or_create_label: getOrCreateLabel,
  
  // Attachment handling
  download_attachment: downloadAttachment,
};