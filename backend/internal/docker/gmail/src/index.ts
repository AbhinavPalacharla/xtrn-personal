import greet from "./tools/greet";
import listEmails from "./tools/list_emails";
import getEmail from "./tools/get_email";
import sendEmail from "./tools/send_email";
import searchEmails from "./tools/search_emails";
import listLabels from "./tools/list_labels";

export default {
  greet,
  list_emails: listEmails,
  get_email: getEmail,
  send_email: sendEmail,
  search_emails: searchEmails,
  list_labels: listLabels,
};