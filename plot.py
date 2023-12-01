import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns

# Load the CSV data into a pandas DataFrame
df = pd.read_csv('results.csv')

# Group by 'Party Setup' and calculate the mean for each metric
avg_times = df.groupby('Party Setup').mean()

# X-axis labels
x_labels = ['3-of-2', '4-of-2', '5-of-2', '6-of-2']

# Set the style of the plots to be 'whitegrid'
sns.set_style('white')

# Set the context to 'paper' which is suited for smaller plots
sns.set_context('paper')

# # Plotting
# plt.figure(figsize=(10, 6))

# Extract Time plot
# plt.plot(x_labels, avg_times['Extract Time (µs)'], color='red', marker='o', label='Extract Time')
sns.lineplot(x=x_labels, y=avg_times['Extract Time (µs)'], color='red', marker='o', label='Extract Time')

# Lookup Time plot
# plt.plot(x_labels, avg_times['Lookup Time (µs)'], color='blue', marker='o', label='Lookup Time')
sns.lineplot(x=x_labels, y=avg_times['Lookup Time (µs)'], color='blue', marker='o', label='Lookup Time')

# Calculate the overall averages for each function
avg_extract_time = avg_times['Extract Time (µs)'].mean()
avg_lookup_time = avg_times['Lookup Time (µs)'].mean()
avg_combined_time = avg_extract_time + avg_lookup_time

# Add the averages to the top of the graph
plt.figtext(0.15, 0.28, f"Extract Filter Avg: {avg_extract_time:.2f}µs", horizontalalignment='left', color='red')
plt.figtext(0.15, 0.25, f"Lookup Avg: {avg_lookup_time:.2f}µs", horizontalalignment='left', color='blue')
plt.figtext(0.15, 0.22, f"Combined Avg: {avg_combined_time:.2f}µs", horizontalalignment='left', color='green')

# Increase the size of the values in the x-axis and y-axis
plt.tick_params(axis='x', labelsize=14)
plt.tick_params(axis='y', labelsize=14)

# Setting labels, title, and legend
plt.xlabel('Party Sets', fontsize=14)
plt.ylabel('Time (µs)', fontsize=14)
plt.title('Average Time vs Party Sets', fontsize=16)
plt.legend(fontsize=12)

# Save the plot as SVG
plt.tight_layout()
plt.savefig('graph.svg', format='svg', dpi=300)
plt.show()
